package pod

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/comm/xterm"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
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

func removePod(ctx context.Context, selectedCluster string, ns string, podName string) {
	// 删除Pod
	kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Pod{}).Name(podName).Namespace(ns).Delete()
}

func cmdLogger(c *gin.Context, cmd string) {
	ns := c.Param("ns")
	podName := c.Param("pod_name")
	containerName := c.Query("container_name")
	selectedCluster := amis.GetSelectedCluster(c)
	cmd = utils.CleanANSISequences(cmd)
	username, role := amis.GetLoginUser(c)
	log := models.ShellLog{
		Cluster:       selectedCluster,
		Command:       cmd,
		Namespace:     ns,
		PodName:       podName,
		ContainerName: containerName,
		UserName:      username,
		Role:          role,
	}
	service.ShellLogService().Add(&log)

}

func Xterm(c *gin.Context) {
	removeAfterExec := c.Query("remove")
	ns := c.Param("ns")
	podName := c.Param("pod_name")
	containerName := c.Query("container_name")
	ctx := amis.GetContextWithUser(c)
	selectedCluster := amis.GetSelectedCluster(c)

	// TODO 转移到kom中，走cb
	var err error
	username := fmt.Sprintf("%s", ctx.Value(constants.JwtUserName))
	roles := fmt.Sprintf("%s", ctx.Value(constants.JwtUserRole))
	clusterRoles, _ := service.UserService().GetClusterRole(selectedCluster, username, roles)

	if len(clusterRoles) == 0 || !(slice.Contain(clusterRoles, constants.RolePlatformAdmin) || slice.Contain(clusterRoles, constants.RoleClusterAdmin) || slice.Contain(clusterRoles, constants.RoleClusterPodExec)) {
		amis.WriteJsonError(c, fmt.Errorf("非管理员,且无exec权限，不能执行Exec命令"))
		return
	}

	if containerName == "" {
		amis.WriteJsonError(c, errors.New("container_name is required"))
		return
	}

	// 使用sync.Once确保清理动作只执行一次
	var cleanupOnce sync.Once
	cleanup := func() {
		if removeAfterExec != "" {
			removePod(ctx, selectedCluster, ns, podName)
		}
	}
	// 确保函数退出时执行清理
	defer cleanupOnce.Do(cleanup)

	// 设置连接超时
	ctx, cancel := context.WithTimeout(ctx, 1*time.Hour)
	defer cancel()

	// 处理信号以确保清理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		cleanupOnce.Do(cleanup)
		cancel()
	}()

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

	cluster := kom.Cluster(selectedCluster).WithContext(ctx)

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
		Param("command", "-c").
		Param("command", "TERM=xterm-256color; export TERM; [ -x /bin/bash ] && ([ -x /usr/bin/script ] && /usr/bin/script -q -c '/bin/bash' /dev/null || exec /bin/bash) || exec /bin/sh").
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
	var outBuffer xterm.SafeBuffer
	var errBuffer xterm.SafeBuffer
	inReader, inWriter := io.Pipe()
	defer inReader.Close()
	defer func() {
		if err := conn.Close(); err != nil {
			cleanupOnce.Do(cleanup)
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
		defer cleanupOnce.Do(cleanup)
		for {
			if err := conn.WriteMessage(websocket.PingMessage, []byte("keepalive")); err != nil {
				klog.V(6).Infof("failed to write ping message")
				cleanupOnce.Do(cleanup)
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
		defer cleanupOnce.Do(cleanup)
		errorCounter := 0
		for {
			// consider the connection closed/errored out so that the socket handler
			// can be terminated - this frees up memory so the service doesn't get
			// overloaded
			if errorCounter > connectionErrorLimit {
				klog.V(6).Infof("connection error limit reached, closing connection")
				cleanupOnce.Do(cleanup)
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
		defer cleanupOnce.Do(cleanup)
		// 创建一个静态缓冲区用于存储命令
		var cmdBuffer bytes.Buffer
		var cmdBufferMutex sync.Mutex
		for {
			// data processing
			messageType, data, err := conn.ReadMessage()
			if err != nil {
				if !connectionClosed {
					cleanupOnce.Do(cleanup)
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
				klog.V(6).Infof("failed to write %d bytes to tty: %v", len(dataBuffer), err)
				continue
			}

			// 使用互斥锁保护 cmdBuffer 的读写操作
			cmdBufferMutex.Lock()
			cmdBuffer.Write(data)
			if bytes.Contains(data, []byte("\r")) {
				// 获取完整命令并去除回车符
				cmd := strings.TrimSuffix(cmdBuffer.String(), "\r")
				// 只有当命令不为空时才记录
				if strings.TrimSpace(cmd) != "" {
					klog.V(8).Infof("收到完整命令: %s", cmd)
					go cmdLogger(c, cmd)
				}
				// 清空缓冲区,准备接收新命令
				cmdBuffer.Reset()
			}
			cmdBufferMutex.Unlock()
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
		cleanupOnce.Do(cleanup)
		return
	}
	// 等待连接关闭或上下文取消
	go func() {
		<-ctx.Done()
		cleanupOnce.Do(cleanup)
		conn.Close()
	}()

	waiter.Wait()
	klog.V(6).Infof("closing conn...")
	connectionClosed = true
	cleanupOnce.Do(cleanup)
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
