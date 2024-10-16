package sse

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func WriteSSE(c *gin.Context, stream io.ReadCloser) {
	defer func() {
		if err := stream.Close(); err != nil {
			// 处理关闭流时的错误
			log.Printf("stream close error:%v", err)
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
