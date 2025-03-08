package chat

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/comm/xterm"
	"github.com/weibaohui/k8m/pkg/service"
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

func GPTShell(c *gin.Context) {

	if !service.AIService().IsEnabled() {
		amis.WriteJsonData(c, gin.H{
			"result": "请先配置开启ChatGPT功能",
		})
		return
	}

	var data ResourceData
	err := c.ShouldBindQuery(&data)
	if err != nil {
		amis.WriteJsonError(c, err)
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

	var outBuffer xterm.SafeBuffer
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

	// chatgpt >> ws
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

				if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
					klog.V(6).Infof("Failed to send stderr message   to xterm.js: %v", err)
					errorCounter++
					return
				} else {
					klog.V(6).Infof("Sent stdout (%d bytes) to xterm.js : %s", len(data), string(data))
					errorCounter = 0
				}

			}

			time.Sleep(100 * time.Millisecond)
			errorCounter = 0
		}
	}()

	// chatgpt << ws
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
			klog.V(6).Infof("received %s (type: %v) message of size %v byte(s) from web ui with key sequence: %v  [%s]", dataType, messageType, dataLength, dataBuffer, string(dataBuffer))

			klog.V(6).Infof("prompt: %s", string(data))

			stream, err := service.ChatService().GetChatStream(string(data))
			if err != nil {
				klog.V(6).Infof(fmt.Sprintf("failed to write %v bytes to tty: %s", len(dataBuffer), err))
				continue
			}

			for {
				response, err := stream.Recv()
				if err != nil {
					if err == io.EOF {
						break
					}
					// 处理其他错误
					continue
				}

				// 发送数据给客户端
				// 写入outBuffer
				outBuffer.Write([]byte(response.Choices[0].Delta.Content))

			}

		}
	}()
	waiter.Wait()
	select {}
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
