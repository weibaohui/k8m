package ai

import (
	"context"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/flag"
	"k8s.io/klog/v2"
)

func (c *OpenAIClient) processThinkFlag(contents ...any) []any {
	cfg := flag.Init()
	if !cfg.Think {
		for i := range contents {
			if txt, ok := contents[i].(string); ok {
				if strings.Contains(txt, "/no_think") {
					continue
				}
				klog.V(6).Infof("关闭  思考  功能 内容[ %s]", txt)
				contents[i] = "/no_think" + txt
			}
		}
	}
	return contents
}

func (c *OpenAIClient) GetCompletion(ctx context.Context, contents ...any) (string, error) {
	contents = c.processThinkFlag(contents...)
	c.fillChatHistory(ctx, contents)

	// Create a completion request
	resp, err := c.client.CreateChatCompletion(ctx,
		openai.ChatCompletionRequest{
			Model:    c.model,
			Messages: c.GetHistory(ctx),
		})
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}
func (c *OpenAIClient) GetCompletionWithTools(ctx context.Context, contents ...any) ([]openai.ToolCall, string, error) {
	contents = c.processThinkFlag(contents...)

	// Create a completion request
	c.fillChatHistory(ctx, contents)
	resp, err := c.client.CreateChatCompletion(ctx,
		openai.ChatCompletionRequest{
			Model:       c.model,
			Messages:    c.GetHistory(ctx),
			Temperature: c.temperature,
			TopP:        c.topP,
			Tools:       c.tools,
		})
	if err != nil {
		return nil, "", err
	}
	return resp.Choices[0].Message.ToolCalls, resp.Choices[0].Message.Content, nil
}

func (c *OpenAIClient) GetStreamCompletion(ctx context.Context, contents ...any) (*openai.ChatCompletionStream, error) {
	contents = c.processThinkFlag(contents...)

	c.fillChatHistory(ctx, contents)
	stream, err := c.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:       c.model,
		Messages:    c.GetHistory(ctx),
		Temperature: c.temperature,
		TopP:        c.topP,
		Stream:      true,
	})
	return stream, err
}
func (c *OpenAIClient) GetStreamCompletionWithTools(ctx context.Context, contents ...any) (*openai.ChatCompletionStream, error) {
	contents = c.processThinkFlag(contents...)

	c.fillChatHistory(ctx, contents)
	stream, err := c.client.CreateChatCompletionStream(ctx, openai.ChatCompletionRequest{
		Model:    c.model,
		Messages: c.GetHistory(ctx),
		Tools:    c.tools,
		Stream:   true,
	})
	klog.V(6).Infof("GetStreamCompletionWithTools 携带 history length: %d", len(c.GetHistory(ctx)))
	klog.V(8).Infof("GetStreamCompletionWithTools c.history: %v", utils.ToJSON(c.GetHistory(ctx)))
	return stream, err
}
