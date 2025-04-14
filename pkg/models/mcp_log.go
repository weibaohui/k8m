package models

import (
	"time"
)

// MCPToolLog MCP工具执行日志
type MCPToolLog struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	ToolName    string    `gorm:"index" json:"tool_name,omitempty"`   // 工具名称
	ServerName  string    `gorm:"index" json:"server_name,omitempty"` // 服务器名称
	Parameters  string    `json:"parameters,omitempty"`               // 执行参数
	Result      string    `json:"result,omitempty"`                   // 执行结果
	Error       string    `json:"error,omitempty"`                    // 错误信息
	ExecuteTime int64     `json:"execute_time,omitempty"`             // 执行时间(毫秒)
	CreatedAt   time.Time `json:"created_at,omitempty"`
	CreatedBy   string    `gorm:"index" json:"created_by,omitempty"`
}
