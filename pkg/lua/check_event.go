package lua

import (
	"time"
)

// CheckEvent 用于记录每次检测的详细信息，包括检测状态、消息、额外上下文、脚本名称、资源类型、描述、命名空间和资源名。
type CheckEvent struct {
	Status     string                 `json:"status"` // 检查状态（如“正常”、“失败”）
	Msg        string                 `json:"msg"`
	Extra      map[string]interface{} `json:"extra,omitempty"`
	ScriptName string                 `json:"scriptName"` // 检测脚本名称
	Kind       string                 `json:"kind"`       // 检查的资源类型
	CheckDesc  string                 `json:"checkDesc"`  // 检查脚本内容描述
	Namespace  string                 `json:"ns"`         // 资源命名空间
	Name       string                 `json:"name"`       // 资源名称
}

type CheckResult struct {
	Name         string
	StartTime    time.Time
	EndTime      time.Time
	LuaRunOutput string
	LuaRunError  error
	Events       []CheckEvent
}
