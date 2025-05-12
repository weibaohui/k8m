package ai

import (
	"sync"

	"github.com/sashabaranov/go-openai"
)

// MemoryService 用于按用户隔离存储和获取对话历史
// 线程安全，适合多用户在线服务场景
// 历史数据以用户名为 key 进行隔离

type memoryService struct {
	mu      sync.RWMutex
	storage map[string][]openai.ChatCompletionMessage // 用户名 -> 对话历史
}

// NewMemoryService 创建 MemoryService 实例
func NewMemoryService() *memoryService {

	return &memoryService{
		storage: make(map[string][]openai.ChatCompletionMessage),
	}
}

// GetUserHistory 获取指定用户的对话历史
func (m *memoryService) GetUserHistory(username string) []openai.ChatCompletionMessage {
	m.mu.RLock()
	defer m.mu.RUnlock()
	history := m.storage[username]
	// 返回副本，避免外部修改
	copied := make([]openai.ChatCompletionMessage, len(history))
	copy(copied, history)
	return copied
}

// AppendUserHistory 向指定用户追加一条历史记录
func (m *memoryService) AppendUserHistory(username string, msg openai.ChatCompletionMessage) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.storage[username] = append(m.storage[username], msg)
}

// ClearUserHistory 清空指定用户的历史记录
func (m *memoryService) ClearUserHistory(username string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.storage, username)
}

func (m *memoryService) SetUserHistory(username string, history []openai.ChatCompletionMessage) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.storage[username] = history
}
