package ai

import (
	"context"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/sashabaranov/go-openai"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/models"
	"k8s.io/klog/v2"
)

// getUsernameFromContext 从 context.Context 提取 JwtUserName 字段的用户名，若获取失败则返回默认用户名
func getUsernameFromContext(ctx context.Context) string {
	val := ctx.Value(constants.JwtUserName)
	username, ok := val.(string)
	if !ok || username == "" {
		return "default_user"
	}
	return username
}

func (c *OpenAIClient) SaveAIHistory(ctx context.Context, contents string) {
	username := getUsernameFromContext(ctx)
	c.memory.AppendUserHistory(username, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: contents,
	})
}

func (c *OpenAIClient) GetHistory(ctx context.Context) []openai.ChatCompletionMessage {
	username := getUsernameFromContext(ctx)
	return c.memory.GetUserHistory(username)
}

func (c *OpenAIClient) ClearHistory(ctx context.Context) error {
	username := getUsernameFromContext(ctx)
	c.memory.ClearUserHistory(username)
	return nil
}
func (c *OpenAIClient) fillChatHistory(ctx context.Context, contents ...any) {
	history := c.GetHistory(ctx)
	for _, content := range contents {
		switch item := content.(type) {
		case string:
			klog.V(2).Infof("Adding user message to history: %v", item)
			history = append(history, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: item,
			})
		case models.MCPToolCallResult:
			klog.V(2).Infof("Adding user message to history: %v", item)
			history = append(history, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: utils.ToJSON(item),
			})
		case []string:
			klog.V(2).Infof("Adding string array to history: %v", item)
			for _, m := range item {
				history = append(history, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleUser,
					Content: m,
				})
			}
		case []models.MCPToolCallResult:
			klog.V(2).Infof("Adding MCPToolCallResult array to history: %v", item)
			for _, m := range item {
				history = append(history, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleUser,
					Content: utils.ToJSON(m),
				})
			}
		case []interface{}:
			for _, m := range item {
				history = append(history, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleUser,
					Content: utils.ToJSON(m),
				})
			}
		default:
			klog.Warningf("Unhandled content type in Send: %T", item)
		}
	}

	// 保留最后 maxHistory 条（含系统提示）
	if c.maxHistory > 0 && int32(len(history)) > c.maxHistory {
		keep := history[len(history)-int(c.maxHistory):]
		history = keep
	}

	system := slice.Filter(history, func(index int, item openai.ChatCompletionMessage) bool {
		if item.Role == openai.ChatMessageRoleSystem {
			return true
		}
		return false
	})

	if len(system) == 0 {
		// 创建系统消息
		sysMsg := openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: sysPrompt,
		}
		// 将系统消息插入到历史记录最前面
		history = append([]openai.ChatCompletionMessage{sysMsg}, history...)
	}
	username := getUsernameFromContext(ctx)
	c.memory.SetUserHistory(username, history)

}
