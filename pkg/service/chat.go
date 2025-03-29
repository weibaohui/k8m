package service

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/sashabaranov/go-openai"
	"k8s.io/klog/v2"
)

type chatService struct {
}

func (c *chatService) GetChatStream(chat string, tools ...openai.Tool) (*openai.ChatCompletionStream, error) {

	client, err := AIService().DefaultClient()

	if err != nil {
		klog.V(6).Infof("获取AI服务错误 : %v\n", err)
		return nil, fmt.Errorf("获取AI服务错误 : %v", err)
	}
	client.SetTools(tools)

	stream, err := client.GetStreamCompletionWithTools(context.TODO(), chat)

	if err != nil {
		klog.V(6).Infof("ChatCompletion error: %v\n", err)
		return nil, err
	}

	return stream, nil

}
func (c *chatService) Chat(chat string) string {
	client, err := AIService().DefaultClient()

	if err != nil {
		klog.V(2).Infof("获取AI服务错误 : %v\n", err)
		return ""
	}

	result, err := client.GetCompletion(context.TODO(), chat)
	if err != nil {
		klog.V(2).Infof("ChatCompletion error: %v\n", err)
		return ""
	}
	return result
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
