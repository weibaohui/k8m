package pod

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/creack/pty"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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

func Xterm(c *gin.Context) {

	// ns := c.Param("ns")
	// podName := c.Param("pod_name")
	// containerName := c.Param("container_name")
	// ctx := c.Request.Context()
	// selectedCluster := amis.GetSelectedCluster(c)

	connectionErrorLimit := 10

	maxBufferSizeBytes := 512
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

	cmd := exec.Command("/bin/bash")
	cmd.Env = os.Environ()
	tty, err := pty.Start(cmd)
	if err != nil {
		message := fmt.Sprintf("failed to start tty: %s", err)
		klog.Errorf(message)
		conn.WriteMessage(websocket.TextMessage, []byte(message))
		return
	}

	defer func() {
		klog.Info("gracefully stopping spawned tty...")
		if err := cmd.Process.Kill(); err != nil {
			klog.Infof("failed to kill process: %s", err)
		}
		if _, err := cmd.Process.Wait(); err != nil {
			klog.Infof("failed to wait for process to exit: %s", err)
		}
		if err := tty.Close(); err != nil {
			klog.Infof("failed to close spawned tty gracefully: %s", err)
		}
		if err := conn.Close(); err != nil {
			klog.Infof("failed to close webscoket connection: %s", err)
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
				klog.Infof("failed to write ping message")
				return
			}
			time.Sleep(keepalivePingTimeout / 2)
			if time.Now().Sub(lastPongTime) > keepalivePingTimeout {
				klog.Infof("failed to get response from ping, triggering disconnect now...")
				waiter.Done()
				return
			}
			klog.Infof("received response from ping successfully")
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
				waiter.Done()
				break
			}
			buffer := make([]byte, maxBufferSizeBytes)
			readLength, err := tty.Read(buffer)
			if err != nil {
				klog.Infof("failed to read from tty: %s", err)
				if err := conn.WriteMessage(websocket.TextMessage, []byte("bye!")); err != nil {
					klog.Infof("failed to send termination message from tty to xterm.js: %s", err)
				}
				waiter.Done()
				return
			}
			if err := conn.WriteMessage(websocket.BinaryMessage, buffer[:readLength]); err != nil {
				klog.Infof("failed to send %v bytes from tty to xterm.js", readLength)
				errorCounter++
				continue
			}
			klog.Infof("sent message of size %v bytes from tty to xterm.js", readLength)
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
					klog.Infof("failed to get next reader: %s", err)
				}
				return
			}
			dataLength := len(data)
			dataBuffer := bytes.Trim(data, "\x00")
			dataType, ok := WebsocketMessageType[messageType]
			if !ok {
				dataType = "unknown"
			}
			klog.Infof("received %s (type: %v) message of size %v byte(s) from xterm.js with key sequence: %v", dataType, messageType, dataLength, dataBuffer)

			// process
			if dataLength == -1 { // invalid
				klog.Infof("failed to get the correct number of bytes read, ignoring message")
				continue
			}

			// handle resizing
			if messageType == websocket.BinaryMessage {
				if dataBuffer[0] == 1 {
					ttySize := &TTYSize{}
					resizeMessage := bytes.Trim(dataBuffer[1:], " \n\r\t\x00\x01")
					if err := json.Unmarshal(resizeMessage, ttySize); err != nil {
						klog.Infof("failed to unmarshal received resize message '%s': %s", string(resizeMessage), err)
						continue
					}
					klog.Infof("resizing tty to use %v rows and %v columns...", ttySize.Rows, ttySize.Cols)
					if err := pty.Setsize(tty, &pty.Winsize{
						Rows: ttySize.Rows,
						Cols: ttySize.Cols,
					}); err != nil {
						klog.Infof("failed to resize tty, error: %s", err)
					}
					continue
				}
			}

			// write to tty
			bytesWritten, err := tty.Write(dataBuffer)
			if err != nil {
				klog.Infof(fmt.Sprintf("failed to write %v bytes to tty: %s", len(dataBuffer), err))
				continue
			}
			klog.Infof("%v bytes written to tty...", bytesWritten)
		}
	}()

	waiter.Wait()
	klog.Infof("closing conn...")
	connectionClosed = true
}
