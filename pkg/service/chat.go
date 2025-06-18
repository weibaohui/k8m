package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sashabaranov/go-openai"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/flag"
	"k8s.io/klog/v2"
)

type chatService struct{}

// getChatStreamBase 是 GetChatStream 和 GetChatStreamWithoutHistory 的通用实现，支持可选的历史清理
// 参数 clearHistory 表示是否在请求前后都清空历史
func (c *chatService) getChatStreamBase(ctx context.Context, chat string, clearHistory bool) (*openai.ChatCompletionStream, error) {
	client, err := AIService().DefaultClient()
	if err != nil {
		klog.V(6).Infof("获取AI服务错误 : %v\n", err)
		return nil, fmt.Errorf("获取AI服务错误 : %v", err)
	}
	defer func() {
		if clearHistory {
			err = client.ClearHistory(ctx)
			if err != nil {
				klog.V(6).Infof("AI服务执行后清空历史失败 : %v\n", err)
			}
		}
	}()
	if clearHistory {
		err = client.ClearHistory(ctx)
		if err != nil {
			klog.V(6).Infof("AI服务执行前清空历史失败 : %v\n", err)
			return nil, fmt.Errorf("AI服务执行前清空历史失败 : %v", err)
		}
	}

	tools := McpService().GetAllEnabledTools()
	klog.V(6).Infof("GPTShell 对话携带tools %d", len(tools))
	client.SetTools(tools)
	stream, err := client.GetStreamCompletionWithTools(ctx, chat)
	if err != nil {
		klog.V(6).Infof("ChatCompletion error: %v\n", err)
		return nil, err
	}
	return stream, nil
}
func (c *chatService) GetChatStream(ctx context.Context, chat string) (*openai.ChatCompletionStream, error) {
	// 仅复用基础方法，不清理历史
	return c.getChatStreamBase(ctx, chat, false)
}
func (c *chatService) GetChatStreamWithoutHistory(ctx context.Context, chat string) (*openai.ChatCompletionStream, error) {
	// 复用基础方法，前后都清理历史
	return c.getChatStreamBase(ctx, chat, true)
}
func (c *chatService) RunOneRound(ctx context.Context, chat string, writer io.Writer) error {
	cfg := flag.Init()

	client, err := AIService().DefaultClient()

	if err != nil {
		klog.V(6).Infof("获取AI服务错误 : %v\n", err)
		return fmt.Errorf("获取AI服务错误 : %v", err)
	}
	tools := McpService().GetAllEnabledTools()
	klog.V(6).Infof("GPTShell 对话携带tools %d", len(tools))

	client.SetTools(tools)

	// currChatContent tracks chat content that needs to be sent
	// to the LLM in each iteration of  the agentic loop below
	var currChatContent []any

	// Set the initial message to start the conversation
	currChatContent = append(currChatContent, chat)

	currentIteration := int32(0)
	maxIterations := cfg.MaxIterations

	for currentIteration < maxIterations {
		klog.Infof("Starting iteration %d", currentIteration)
		if len(currChatContent) == 0 {
			klog.V(6).Infof("No content to send to LLM")
			return nil
		}
		klog.V(6).Infof("Sending to LLM: %v", utils.ToJSON(currChatContent))
		stream, err := client.GetStreamCompletionWithTools(ctx, currChatContent...)
		// Clear our "response" now that we sent the last response
		if err != nil {
			klog.V(6).Infof("ChatCompletion error: %v\n", err)
			return err
		}
		currChatContent = nil

		var toolCallBuffer []openai.ToolCall
		var respBuffer []string
		for {
			response, recvErr := stream.Recv()
			if recvErr != nil {
				if recvErr == io.EOF {
					break
				}
				klog.V(6).Infof("stream Recv error:%v", recvErr)
				// 处理其他错误
				continue
			}

			// 设置了工具
			if len(response.Choices) > 0 {
				for _, choice := range response.Choices {
					// 大模型选择了执行工具
					// 解析当前的ToolCalls
					var currentCalls []openai.ToolCall
					if err = json.Unmarshal([]byte(utils.ToJSON(choice.Delta.ToolCalls)), &currentCalls); err == nil {
						toolCallBuffer = append(toolCallBuffer, currentCalls...)
					}

					// 当收到空的ToolCalls时，表示一个完整的ToolCall已经接收完成
					if len(choice.Delta.ToolCalls) == 0 && len(toolCallBuffer) > 0 {
						// 合并并处理完整的ToolCall
						mergedCalls := MergeToolCalls(toolCallBuffer)

						klog.V(6).Infof("合并最终ToolCalls: %v", utils.ToJSON(mergedCalls))

						// 使用合并后的ToolCalls执行操作

						results := McpService().Host().ExecTools(ctx, mergedCalls)
						for _, r := range results {
							currChatContent = append(currChatContent, gin.H{
								"type": "执行结果",
								"raw":  r,
							})
							_, _ = writer.Write([]byte(utils.ToJSON(r)))
						}

						// 清空缓冲区
						toolCallBuffer = nil
					}
				}

			}

			// 发送数据给客户端
			// 写入outBuffer
			content := response.Choices[0].Delta.Content
			respBuffer = append(respBuffer, content)

			_, _ = writer.Write([]byte(content))
		}
		respAll := strings.Join(respBuffer, "")
		if strings.TrimSpace(respAll) != "" {
			client.SaveAIHistory(ctx, respAll)
		}
		respBuffer = []string{}
		err = stream.Close()
		if err != nil {
			klog.V(6).Infof("stream close error:%v", err)
		}
		klog.V(6).Infof("stream close ")
		currentIteration++
	}

	if currentIteration == maxIterations {
		// If we've reached the maximum number of iterations
		klog.Infof("Max iterations %d reached", maxIterations)
		return fmt.Errorf("max iterations %d reached", maxIterations)
	}
	return nil
}
func (c *chatService) Chat(ctx *gin.Context, chat string) string {
	ctxInst := amis.GetContextWithUser(ctx)
	return c.ChatWithCtx(ctxInst, chat)
}
func (c *chatService) ChatWithCtx(ctx context.Context, chat string) string {
	client, err := AIService().DefaultClient()

	if err != nil {
		klog.V(2).Infof("获取AI服务错误 : %v\n", err)
		return ""
	}

	result, err := client.GetCompletion(ctx, chat)
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

// MergeToolCalls 合并多个分段接收的 ToolCall 数据，生成完整的 ToolCall 切片。
// 适用于将流式返回的部分 ToolCall 信息按索引聚合为完整的调用记录。
//
// 返回合并后的 ToolCall 切片。
func MergeToolCalls(toolCalls []openai.ToolCall) []openai.ToolCall {
	mergedCalls := make(map[int]*openai.ToolCall)

	for _, call := range toolCalls {
		if call.Index == nil {
			continue
		}
		idx := *call.Index
		if existing, ok := mergedCalls[idx]; ok {
			// 合并现有数据
			if call.ID != "" {
				existing.ID = call.ID
			}
			if call.Type != "" {
				existing.Type = call.Type
			}
			if call.Function.Name != "" {
				existing.Function.Name = call.Function.Name
			}
			if call.Function.Arguments != "" {
				existing.Function.Arguments += call.Function.Arguments
			}
		} else {
			// 创建新的ToolCall
			copyCall := call
			mergedCalls[idx] = &copyCall
		}
	}

	// 转换为切片
	result := make([]openai.ToolCall, 0, len(mergedCalls))
	for _, call := range mergedCalls {
		result = append(result, *call)
	}
	return result
}
