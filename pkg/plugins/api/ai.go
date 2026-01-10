package api

import (
	"context"
	"sync/atomic"
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

// noopAIChat 为默认的空实现，保证在未注册真实实现时也不会产生空指针。
type noopAIChat struct{}

func (noopAIChat) Chat(ctx context.Context, prompt string) (string, error) {
	return "AI插件未开启", nil
}

func (noopAIChat) ChatNoHistory(ctx context.Context, prompt string) (string, error) {
	return "AI插件未开启", nil
}

// noopAIConfig 为默认的空实现，提供安全的配置访问。
type noopAIConfig struct{}

func (noopAIConfig) AnySelect() bool {
	return false
}

func (noopAIConfig) FloatingWindow() bool {
	return false
}

var (
	aiChatVal   atomic.Value // 保存 AIChat 实现，始终为非 nil
	aiConfigVal atomic.Value // 保存 AIConfig 实现，始终为非 nil
)

type aiChatHolder struct {
	chat AIChat
}

type aiConfigHolder struct {
	cfg AIConfig
}

func init() {
	aiChatVal.Store(&aiChatHolder{chat: noopAIChat{}})
	aiConfigVal.Store(&aiConfigHolder{cfg: noopAIConfig{}})
}

// AIChatService 返回当前生效的 AIChat 实现，始终非 nil。
func AIChatService() AIChat {
	return aiChatVal.Load().(*aiChatHolder).chat
}

// AIConfigService 返回当前生效的 AIConfig 实现，始终非 nil。
func AIConfigService() AIConfig {
	return aiConfigVal.Load().(*aiConfigHolder).cfg
}

// RegisterAI 在运行期注册或切换 AI 能力实现。
// 传入 nil 时自动回退为 noop 实现，保证始终非 nil。
func RegisterAI(chatImpl AIChat, cfgImpl AIConfig) {
	if chatImpl == nil {
		chatImpl = noopAIChat{}
	}
	if cfgImpl == nil {
		cfgImpl = noopAIConfig{}
	}

	aiChatVal.Store(&aiChatHolder{chat: chatImpl})
	aiConfigVal.Store(&aiConfigHolder{cfg: cfgImpl})
}

// UnregisterAI 在运行期取消注册 AI 能力，实现回退为 noop。
func UnregisterAI() {
	aiChatVal.Store(&aiChatHolder{chat: noopAIChat{}})
	aiConfigVal.Store(&aiConfigHolder{cfg: noopAIConfig{}})
}
