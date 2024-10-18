package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/sashabaranov/go-openai"
	"k8s.io/klog/v2"
)

var model = "Qwen/Qwen2.5-Coder-7B-Instruct"

type ChatService struct {
}

func (c *ChatService) GetChatStream(chat string) (*http.Response, error) {
	key, apiURL, enable := getChatGPTAuth()
	if !enable {
		return nil, fmt.Errorf("chatGPT not enable")
	}

	// url := "https://api.siliconflow.cn/v1/chat/completions"
	url := fmt.Sprintf("%s/chat/completions", apiURL)

	// 构建请求体
	payload := map[string]interface{}{
		"model": model,
		"messages": []map[string]string{
			{
				"role":    "user",
				"content": chat,
			},
		},
		"stream": true,
	}

	// 将请求体编码为JSON
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		klog.V(2).Infof("Error marshaling JSON:%v\n", err)
		return nil, err
	}

	// 设置请求头
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		klog.V(2).Infof("Error creating request:%v\n", err)
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", key))

	// 执行请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		klog.V(2).Infof("Error making request:%v\n\n", err)
		return nil, err
	}
	return resp, err
}
func (c *ChatService) Chat(chat string) string {
	apiKey, apiURL, enable := getChatGPTAuth()
	if !enable {
		return ""
	}
	// 初始化OpenAI客户端
	cfg := openai.DefaultConfig(apiKey)
	cfg.BaseURL = apiURL
	openaiClient := openai.NewClientWithConfig(cfg)

	resp, err := openaiClient.CreateChatCompletion(
		context.TODO(),
		openai.ChatCompletionRequest{
			Model: model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: chat,
				},
			},
		},
	)

	if err != nil {
		klog.V(2).Infof("ChatCompletion error: %v\n", err)
		return ""
	}

	result := resp.Choices[0].Message.Content
	return result
}

func getChatGPTAuth() (apiKey string, apiURL string, enable bool) {
	// 从环境变量读取OpenAI API Key和API URL
	apiKey = os.Getenv("OPENAI_API_KEY")
	apiURL = os.Getenv("OPENAI_API_URL")
	if apiKey != "" && apiURL != "" {
		enable = true
	}
	return
}
