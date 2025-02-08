package pod

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
	"k8s.io/apimachinery/pkg/util/httpstream"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/klog/v2"
)

var WebsocketMessageType = map[int]string{
	websocket.BinaryMessage: "binary",
	websocket.TextMessage:   "text",
	websocket.CloseMessage:  "close",
	websocket.PingMessage:   "ping",
	websocket.PongMessage:   "pong",
}

type TTYSize struct {
	Cols uint16 `json:"cols"`
	Rows uint16 `json:"rows"`
	X    uint16 `json:"x"`
	Y    uint16 `json:"y"`
}

// TerminalSizeQueue 维护 TTY 终端大小
type TerminalSizeQueue struct {
	sync.Mutex
	sizes []remotecommand.TerminalSize
}

func (t *TerminalSizeQueue) Next() *remotecommand.TerminalSize {
	t.Lock()
	defer t.Unlock()
	if len(t.sizes) == 0 {
		return nil
	}
	size := t.sizes[len(t.sizes)-1]
	t.sizes = t.sizes[:len(t.sizes)-1]
	return &size
}

func (t *TerminalSizeQueue) Push(cols, rows uint16) {
	t.Lock()
	defer t.Unlock()
	t.sizes = append(t.sizes, remotecommand.TerminalSize{Width: cols, Height: rows})
}

func Xterm(c *gin.Context) {

	ns := c.Param("ns")
	podName := c.Param("pod_name")
	containerName := c.Query("container_name")
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	if containerName == "" {
		amis.WriteJsonError(c, errors.New("container_name is required"))
		return
	}

	connectionErrorLimit := 10

	keepalivePingTimeout := 20 * time.Second

	// 定义 WebSocket 升级器
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// 允许所有来源
			return true
		},
	}

	// 将 HTTP 连接升级为 WebSocket 连接
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		klog.Errorf("WebSocket Upgrade Error:%v", err)
		return
	}
	defer conn.Close()
	klog.V(6).Infof("ws Client connected")

	cluster := kom.Cluster(selectedCluster)

	// 创建 TTY 终端大小管理队列
	sizeQueue := &TerminalSizeQueue{}
	// 定义 Kubernetes Exec 请求
	req := cluster.Client().CoreV1().RESTClient().
		Post().
		Resource("pods").
		Namespace(ns).
		Name(podName).
		SubResource("exec").
		Param("container", containerName).
		Param("command", "/bin/sh").
		Param("tty", "true").
		Param("stdin", "true").
		Param("stdout", "true").
		Param("stderr", "true")

	// 创建 WebSocket -> Pod 交互的 Executor
	exec, err := createExecutor(req.URL(), cluster.RestConfig())
	if err != nil {
		klog.Errorf("Failed to create SPDYExecutor: %v", err)
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Error creating executor: %v", err)))
		return
	}

	// 用于传输数据
	// var inBuffer SafeBuffer
	var outBuffer SafeBuffer
	var errBuffer SafeBuffer
	inReader, inWriter := io.Pipe()
	defer inReader.Close()
	defer func() {
		if err := conn.Close(); err != nil {
			klog.V(6).Infof("failed to close webscoket connection: %s", err)
		}
	}()

	var connectionClosed bool
	var waiter sync.WaitGroup
	waiter.Add(1)

	// this is a keep-alive loop that ensures connection does not hang-up itself
	lastPongTime := time.Now()
	conn.SetPongHandler(func(msg string) error {
		lastPongTime = time.Now()
		return nil
	})
	go func() {
		for {
			if err := conn.WriteMessage(websocket.PingMessage, []byte("keepalive")); err != nil {
				klog.V(6).Infof("failed to write ping message")
				return
			}
			time.Sleep(keepalivePingTimeout / 2)
			if time.Now().Sub(lastPongTime) > keepalivePingTimeout {
				klog.V(6).Infof("failed to get response from ping, triggering disconnect now...")
				waiter.Done()
				return
			}
			klog.V(6).Infof("received response from ping successfully")
		}
	}()

	// tty >> xterm.js
	go func() {
		errorCounter := 0
		for {
			// consider the connection closed/errored out so that the socket handler
			// can be terminated - this frees up memory so the service doesn't get
			// overloaded
			if errorCounter > connectionErrorLimit {
				klog.V(6).Infof("connection error limit reached, closing connection")
				waiter.Done()
				break
			}

			if outBuffer.Len() > 0 {
				data := outBuffer.Bytes()
				outBuffer.Reset()
				klog.V(6).Infof("Received stdout (%d bytes): %q", len(data), string(data))

				if err := conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
					klog.V(6).Infof("Failed to send stderr message   to xterm.js: %v", err)
					errorCounter++
					return
				} else {
					klog.V(6).Infof("Sent stdout (%d bytes) to xterm.js : %s", len(data), string(data))
					errorCounter = 0
				}

			}
			if errBuffer.Len() > 0 {
				data := errBuffer.Bytes()
				errBuffer.Reset()
				klog.V(6).Infof("Received stderr (%d bytes): %q", len(data), string(data))
				if err := conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
					klog.V(6).Infof("Failed to send stderr message   to xterm.js: %v", err)
					errorCounter++
					return
				}
			}
			time.Sleep(100 * time.Millisecond)
			errorCounter = 0
		}
	}()

	// tty << xterm.js
	go func() {
		for {
			// data processing
			messageType, data, err := conn.ReadMessage()
			if err != nil {
				if !connectionClosed {
					klog.V(6).Infof("failed to get next reader: %s", err)
				}
				return
			}
			dataLength := len(data)
			dataBuffer := bytes.Trim(data, "\x00")
			dataType, ok := WebsocketMessageType[messageType]
			if !ok {
				dataType = "unknown"
			}
			klog.V(6).Infof("received %s (type: %v) message of size %v byte(s) from xterm.js with key sequence: %v  [%s]", dataType, messageType, dataLength, dataBuffer, string(dataBuffer))

			// process
			if dataLength == -1 { // invalid
				klog.V(6).Infof("failed to get the correct number of bytes read, ignoring message")
				continue
			}

			// handle resizing
			if messageType == websocket.BinaryMessage {
				if dataBuffer[0] == 1 {
					ttySize := &TTYSize{}
					resizeMessage := bytes.Trim(dataBuffer[1:], " \n\r\t\x00\x01")
					if err := json.Unmarshal(resizeMessage, ttySize); err != nil {
						klog.V(6).Infof("failed to unmarshal received resize message '%s': %s", string(resizeMessage), err)
						continue
					}
					klog.V(6).Infof("resizing tty to use %v rows and %v columns...", ttySize.Rows, ttySize.Cols)

					sizeQueue.Push(ttySize.Cols, ttySize.Rows)
					continue
				}
			}

			// write to tty
			// 普通输入
			bytesWritten, err := inWriter.Write(data)
			if err != nil {
				klog.V(6).Infof(fmt.Sprintf("failed to write %v bytes to tty: %s", len(dataBuffer), err))
				continue
			}
			klog.V(6).Infof("Wrote %d bytes to inBuffer: %q", bytesWritten, string(data))
		}
	}()

	// 执行命令
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:             inReader,
		Stdout:            &outBuffer,
		Stderr:            &errBuffer,
		Tty:               true,
		TerminalSizeQueue: sizeQueue, // 传递 TTY 尺寸管理队列
	})
	if err != nil {
		klog.Errorf("Failed to execute command in pod: %v", err)
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("Execution error: %v", err)))
		return
	}
	waiter.Wait()
	klog.V(6).Infof("closing conn...")
	connectionClosed = true
}

func createExecutor(url *url.URL, config *rest.Config) (remotecommand.Executor, error) {

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", url)
	if err != nil {
		return nil, err
	}
	// Fallback executor is default, unless feature flag is explicitly disabled.
	// WebSocketExecutor must be "GET" method as described in RFC 6455 Sec. 4.1 (page 17).
	websocketExec, err := remotecommand.NewWebSocketExecutor(config, "GET", url.String())
	if err != nil {
		return nil, err
	}
	exec, err = remotecommand.NewFallbackExecutor(websocketExec, exec, func(err error) bool {
		return httpstream.IsUpgradeFailure(err) || httpstream.IsHTTPSProxyError(err)
	})
	if err != nil {
		return nil, err
	}
	return exec, nil
}
