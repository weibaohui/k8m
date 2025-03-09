package sse

import (
	"bufio"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sashabaranov/go-openai"
	"k8s.io/klog/v2"
)

func WriteSSE(c *gin.Context, stream io.ReadCloser) {
	defer func() {
		if err := stream.Close(); err != nil {
			// 处理关闭流时的错误
			klog.V(6).Infof("stream close error:%v", err)
		}
	}()

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.WriteHeader(http.StatusOK)

	// 逐行读取日志并发送到 Channel
	reader := bufio.NewReader(stream)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			// 处理读取错误，向客户端发送错误消息
			c.SSEvent("error", fmt.Sprintf("Error reading stream: %v", err))
			c.Writer.Flush()
			break
		}
		// 发送 SSE 消息
		c.SSEvent("message", line)
		// 刷新输出缓冲区
		c.Writer.Flush()
	}
}
func WriteSSEWithChannel(c *gin.Context, logCh <-chan string, done chan struct{}) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.WriteHeader(http.StatusOK)

	for {
		select {
		case message, ok := <-logCh:
			if !ok {
				return
			}
			if message == ":heartbeat" {
				c.SSEvent("heartbeat", "")
			} else {
				c.SSEvent("message", message)
			}
			c.Writer.Flush()
		case <-c.Request.Context().Done():
			close(done) // 停止数据库查询
			return
		}
	}
}

func WriteWebSocketChatCompletionStream(c *gin.Context, stream *openai.ChatCompletionStream) {
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

	defer func() {
		if err := stream.Close(); err != nil {
			// 处理关闭流时的错误
			klog.V(6).Infof("stream close error:%v", err)
		}
		klog.V(6).Infof("stream close ")
	}()

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
		conn.WriteJSON(gin.H{
			"data": string(response.Choices[0].Delta.Content),
		})
	}

}

func WriteSSEChatCompletionStream(c *gin.Context, stream *openai.ChatCompletionStream) {
	defer func() {
		if err := stream.Close(); err != nil {
			// 处理关闭流时的错误
			klog.V(6).Infof("stream close error:%v", err)
		}
	}()

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.WriteHeader(http.StatusOK)

	for {
		response, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}
			// 处理其他错误
			continue
		}
		// 发送 SSE 消息
		c.SSEvent("message", response.Choices[0].Delta.Content)
		// 刷新输出缓冲区
		c.Writer.Flush()
	}

}
