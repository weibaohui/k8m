package config

import (
	"crypto/md5"
	"fmt"
	"time"
)

// Event 表示一个Kubernetes事件
type Event struct {
	ID        int64     `json:"id" db:"id"`
	EvtKey    string    `json:"evt_key" db:"evt_key"`       // 事件唯一标识符
	Type      string    `json:"type" db:"type"`             // 事件类型 (Normal/Warning)
	Reason    string    `json:"reason" db:"reason"`         // 事件原因
	Level     string    `json:"level" db:"level"`           // 事件级别
	Namespace string    `json:"namespace" db:"namespace"`   // 命名空间
	Name      string    `json:"name" db:"name"`             // 资源名称
	Message   string    `json:"message" db:"message"`       // 事件消息
	Timestamp time.Time `json:"timestamp" db:"timestamp"`   // 事件发生时间
	Processed bool      `json:"processed" db:"processed"`   // 是否已处理
	Attempts  int       `json:"attempts" db:"attempts"`     // 处理重试次数
	CreatedAt time.Time `json:"created_at" db:"created_at"` // 创建时间
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"` // 更新时间
}

// GenerateEvtKey 生成事件唯一标识符
// 格式：namespace|kind|name|reason|hash(message)
func GenerateEvtKey(namespace, kind, name, reason, message string) string {
	hash := md5.Sum([]byte(message))
	hashStr := fmt.Sprintf("%x", hash)
	return fmt.Sprintf("%s|%s|%s|%s|%s", namespace, kind, name, reason, hashStr[:8])
}

// IsWarning 判断是否为警告类型事件
func (e *Event) IsWarning() bool {
	return e.Type == "Warning"
}

// IsNormal 判断是否为正常类型事件
func (e *Event) IsNormal() bool {
	return e.Type == "Normal"
}

// ShouldProcess 判断事件是否应该被处理
func (e *Event) ShouldProcess() bool {
	return !e.Processed && e.IsWarning()
}
