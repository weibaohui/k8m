package lua

import (
	"time"

	"github.com/weibaohui/k8m/pkg/models"
)

// CheckEvent 用于记录每次检测的详细信息，包括检测状态、消息、额外上下文、脚本名称、资源类型、描述、命名空间和资源名。
type CheckEvent struct {
	Status     string         `json:"status"` // 检查状态（如“正常”、“失败”）
	Msg        string         `json:"msg"`
	Extra      map[string]any `json:"extra,omitempty"`
	ScriptName string         `json:"scriptName"` // 检测脚本名称
	Kind       string         `json:"kind"`       // 检查的资源类型
	CheckDesc  string         `json:"checkDesc"`  // 检查脚本内容描述
	Namespace  string         `json:"ns"`         // 资源命名空间
	Name       string         `json:"name"`       // 资源名称
}

type CheckResult struct {
	Name         string
	StartTime    time.Time
	EndTime      time.Time
	LuaRunOutput string
	LuaRunError  error
	Events       []CheckEvent
}

// SummaryMsg 巡检记录汇总信息结构体
// 用于替代原来的 map[string]any 返回类型，提供类型安全和更好的性能
type SummaryMsg struct {
	RecordDate        string                          `json:"record_date"`        // 巡检记录日期（本地时间格式）
	RecordID          uint                            `json:"record_id"`          // 巡检记录ID
	ScheduleID        *uint                           `json:"schedule_id"`        // 巡检计划ID
	ScheduleName      string                          `json:"schedule_name"`      // 巡检计划名称
	Cluster           string                          `json:"cluster"`            // 集群名称
	TotalRules        int                             `json:"total_rules"`        // 总规则数
	FailedCount       int                             `json:"failed_count"`       // 失败数量
	FailedList        []*models.InspectionCheckEvent  `json:"failed_list"`        // 失败事件列表
	AIEnabled         bool                            `json:"ai_enabled"`         // 是否启用AI汇总
	AIPromptTemplate  string                          `json:"ai_prompt_template"` // AI提示模板
}
