package pod

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/creack/pty"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/controller/sse"
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func StreamLogs(c *gin.Context) {

	ns := c.Param("ns")
	podName := c.Param("pod_name")
	containerName := c.Param("container_name")
	selector := fmt.Sprintf("metadata.name=%s", podName)
	StreamPodLogsBySelector(c, ns, containerName, metav1.ListOptions{
		FieldSelector: selector,
	})
}
func StreamPodLogsBySelector(c *gin.Context, ns string, containerName string, options metav1.ListOptions) {
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	var pods []v1.Pod
	err := kom.Cluster(selectedCluster).Resource(&v1.Pod{}).Namespace(ns).List(&pods, options).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	if len(pods) != 1 {
		amis.WriteJsonError(c, errors.New("pod 数量过多"))
		return
	}

	var podName = pods[0].GetName()
	logOpt, err := BindPodLogOptions(c, containerName)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	podService := service.PodService()
	stream, err := podService.StreamPodLogs(ctx, selectedCluster, ns, podName, logOpt)

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	sse.WriteSSE(c, stream)
}
func DownloadLogs(c *gin.Context) {

	ns := c.Param("ns")
	podName := c.Param("pod_name")
	containerName := c.Param("container_name")
	selector := fmt.Sprintf("metadata.name=%s", podName)
	DownloadPodLogsBySelector(c, ns, containerName, metav1.ListOptions{FieldSelector: selector})
}

func WsExec(c *gin.Context) {

	ns := c.Param("ns")
	podName := c.Param("pod_name")
	containerName := c.Param("container_name")
	cmd := c.Query("cmd")
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	if cmd == "" {
		amis.WriteJsonError(c, fmt.Errorf("执行命令为空"))
		return
	}

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

	klog.V(6).Infof("cmd=%s\n", cmd)

	cb := func(data []byte) error {
		// 发送数据给客户端
		conn.WriteJSON(gin.H{
			"data": string(data),
		})
		return nil
	}
	err = kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.Pod{}).
		Namespace(ns).
		Name(podName).Ctl().Pod().
		ContainerName(containerName).
		Command("sh", "-c", cmd).
		StreamExecute(cb, cb).Error

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	select {}
}

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
func Exec(c *gin.Context) {
	ns := c.Param("ns")
	podName := c.Param("pod_name")
	containerName := c.Param("container_name")
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Second*5))
	defer cancel()
	// 初始化结构体实例
	var payload struct {
		Command string `json:"cmd"`
	}

	// 反序列化 JSON 数据到结构体
	if err := c.ShouldBindJSON(&payload); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	var humanCommand string
	chatService := service.ChatService()
	humanCommand = payload.Command
	if chatService.IsEnabled() {
		prompt := fmt.Sprintf("请根据用户描述，给出最合适的一条命令。第一步，给出命令，第二步，检查命令是否为单行单个命令。请务必注意，只给出一条命令。请不要使用top、tail -f等流式输出的命令，请要不使用tzdate等交互性的命令。只能使用输入命令，紧接着输出完整返回的命令。请不要做任何解释。最终的代码一定、务必、必须用```bash\n命令写这里\n```包裹起来\n以下为用户的要求:\n%s", strings.TrimPrefix(payload.Command, "#"))
		aiCmd := chatService.Chat(prompt)
		cleanCmd := chatService.CleanCmd(aiCmd)
		klog.V(4).Infof("\n用户输入:\t%s\nprompt:\t%s\nAI返回:\t%s\n提取命令:\t%s\n", payload.Command, prompt, aiCmd, cleanCmd)
		if cleanCmd != "" {
			payload.Command = cleanCmd
		}
	} else {
		// 未开启，那么删除掉#,提高容错处理
		humanCommand = strings.TrimPrefix(payload.Command, "#")
		payload.Command = humanCommand
	}

	var result []byte
	err := kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.Pod{}).
		Namespace(ns).
		Name(podName).Ctl().Pod().
		ContainerName(containerName).Command("sh", "-c", payload.Command).Execute(&result).Error

	if err != nil {
		amis.WriteJsonData(c, gin.H{
			"result":        fmt.Sprintf("%v", err.Error()),
			"human_command": humanCommand,
			"last_command":  payload.Command,
		})
		return
	}
	amis.WriteJsonData(c, gin.H{
		"result":        string(result),
		"human_command": humanCommand,
		"last_command":  payload.Command,
	})

}
func DownloadPodLogsBySelector(c *gin.Context, ns string, containerName string, options metav1.ListOptions) {
	selectedCluster := amis.GetSelectedCluster(c)

	ctx := c.Request.Context()
	var pods []v1.Pod
	err := kom.Cluster(selectedCluster).Resource(&v1.Pod{}).Namespace(ns).List(&pods, options).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	if len(pods) != 1 {
		amis.WriteJsonError(c, errors.New("pod 数量过多"))
		return
	}

	var podName = pods[0].GetName()
	logOpt, err := BindPodLogOptions(c, containerName)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	logOpt.Follow = false

	podService := service.PodService()
	stream, err := podService.StreamPodLogs(ctx, selectedCluster, ns, podName, logOpt)

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	sse.DownloadLog(c, logOpt, stream)
}
func Usage(c *gin.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	usage := kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.Pod{}).
		Namespace(ns).
		Name(name).
		Ctl().Pod().ResourceUsageTable()
	amis.WriteJsonData(c, usage)
}
