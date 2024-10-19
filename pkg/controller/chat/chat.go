package chat

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/service"
	"k8s.io/klog/v2"
)

func Chat(c *gin.Context) {
	q := c.Query("q")
	chatService := service.ChatService{}
	result := chatService.Chat(q)
	amis.WriteJsonData(c, result)
}
func Sse(c *gin.Context) {
	q := c.Query("q")
	chatService := service.ChatService{}
	resp, err := chatService.GetChatStream(q)
	if err != nil {
		klog.V(2).Infof("Error making request:%v\n\n", err)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			klog.V(6).Infof("Body close error:%v\n", err)
		}
	}(resp.Body)

	// 检查响应状态码
	if resp.StatusCode == http.StatusOK {
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			line = strings.TrimPrefix(line, "data:")
			line = strings.TrimSpace(line)
			if line != "" && line != "[DONE]" {
				chatResult := &openai.ChatCompletionStreamResponse{}
				err := json.Unmarshal([]byte(line), chatResult)
				if err != nil {
					c.SSEvent("message", fmt.Sprintf("json error:%v\n%s", err, line))
					continue
				}
				if len(chatResult.Choices) > 0 {
					message := chatResult.Choices[0].Delta.Content
					c.SSEvent("message", message)
				}
			}

		}
		if err := scanner.Err(); err != nil {
			c.SSEvent("message", fmt.Sprintf("Error reading response:%v", err))
		}
	} else {
		fmt.Println("Request failed with status code:", resp.StatusCode)
	}

}
