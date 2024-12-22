package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/sashabaranov/go-openai"
	"k8s.io/klog/v2"
)

type chatService struct {
	model  string
	apiKey string
	apiUrl string
}

func (c *chatService) SetVars(apikey, apiUrl, model string) {
	c.model = model
	c.apiUrl = apiUrl
	c.apiKey = apikey
}

// Deprecated 获取stream
func (c *chatService) GetChatStream1(chat string) (*http.Response, error) {
	key, apiURL, enable := c.getChatGPTAuth()
	if !enable {
		return nil, fmt.Errorf("chatGPT not enable")
	}

	// url := "https://api.siliconflow.cn/v1/chat/completions"
	url := fmt.Sprintf("%s/chat/completions", apiURL)

	// 构建请求体
	payload := map[string]interface{}{
		"model": c.model,
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
func (c *chatService) GetChatStream(chat string) (*openai.ChatCompletionStream, error) {
	apiKey, apiURL, enable := c.getChatGPTAuth()
	if !enable {
		return nil, fmt.Errorf("chatGPT not enable")
	}
	// 初始化OpenAI客户端
	cfg := openai.DefaultConfig(apiKey)
	cfg.BaseURL = apiURL
	openaiClient := openai.NewClientWithConfig(cfg)

	stream, err := openaiClient.CreateChatCompletionStream(context.Background(), openai.ChatCompletionRequest{
		Model: c.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: chat,
			},
		},
		Stream: true,
	})

	if err != nil {
		klog.V(2).Infof("ChatCompletion error: %v\n", err)
		return nil, err
	}

	return stream, nil

}
func (c *chatService) Chat(chat string) string {
	apiKey, apiURL, enable := c.getChatGPTAuth()
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
			Model: c.model,
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

func (c *chatService) getChatGPTAuth() (apiKey string, apiURL string, enable bool) {
	// 从环境变量读取OpenAI API Key和API URL
	// 环境变量优先
	apiKey = os.Getenv("OPENAI_API_KEY")
	apiURL = os.Getenv("OPENAI_API_URL")
	if apiKey == "" && apiURL == "" {
		apiKey = c.apiKey
		apiURL = c.apiUrl
	}
	if apiKey != "" && apiURL != "" {
		enable = true
	}

	c.apiKey = apiKey
	c.apiUrl = apiURL
	return
}
func (c *chatService) IsEnabled() bool {
	_, _, enable := c.getChatGPTAuth()
	return enable
}

// CleanCmd 提取Markdown包裹的命令正文
func (c *chatService) CleanCmd(cmd string) string {
	// 去除首尾空白字符
	cmd = strings.TrimSpace(cmd)

	// 正则表达式匹配三个反引号包裹的命令，忽略语言标记
	reCommand := regexp.MustCompile("(?s)```(?:bash|sh|zsh|cmd|powershell)?\\s+(.*?)\\s+```")
	match := reCommand.FindStringSubmatch(cmd)

	// 如果找到匹配的命令正文，返回去除前后空格的结果
	if len(match) > 1 {
		return strings.TrimSpace(match[1])
	}

	return ""
}
