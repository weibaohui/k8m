package api

import (
	"context"
	"sync"
)

// AIChat 抽象 AI 聊天能力，对调用方隐藏具体插件实现和内部逻辑。
type AIChat interface {
	// Chat 使用带用户上下文的对话能力，可能带历史记录。
	Chat(ctx context.Context, prompt string) (string, error)
	// ChatNoHistory 使用不带历史记录的对话能力，适合一次性问题。
	ChatNoHistory(ctx context.Context, prompt string) (string, error)
}

// AIConfig 提供只读的 AI 配置视图，避免外部直接依赖具体 aiService 结构。
type AIConfig interface {
	AnySelect() bool
	FloatingWindow() bool
}

var (
	aiChatImpl   AIChat
	aiConfigImpl AIConfig
	aiMu         sync.RWMutex
)

// RegisterAIChat 由 AI 插件在生命周期中调用，用于注册聊天能力实现。
func RegisterAIChat(impl AIChat) {
	aiMu.Lock()
	defer aiMu.Unlock()
	aiChatImpl = impl
}

// RegisterAIConfig 由 AI 插件在生命周期中调用，用于注册配置视图实现。
func RegisterAIConfig(impl AIConfig) {
	aiMu.Lock()
	defer aiMu.Unlock()
	aiConfigImpl = impl
}

// AIChatService 返回已注册的 AIChat 实现，未注册时返回 nil。
func AIChatService() AIChat {
	aiMu.RLock()
	defer aiMu.RUnlock()
	return aiChatImpl
}

// AIConfigService 返回已注册的 AIConfig 实现，未注册时返回 nil。
func AIConfigService() AIConfig {
	aiMu.RLock()
	defer aiMu.RUnlock()
	return aiConfigImpl
}
