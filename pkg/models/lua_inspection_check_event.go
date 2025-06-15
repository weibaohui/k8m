package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

// InspectionCheckEvent  用于记录每次检测的详细信息，包括检测状态、消息、额外上下文、脚本名称、资源类型、描述、命名空间和资源名。
type InspectionCheckEvent struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	RecordID    uint      `json:"record_id"`                        // 关联的巡检执行记录ID
	EventStatus string    `json:"event_status"`                     // 事件状态（如“正常”、“失败”）
	EventMsg    string    `json:"event_msg"`                        // 事件消息
	Extra       string    `gorm:"type:text" json:"extra,omitempty"` // 额外上下文
	ScriptName  string    `json:"script_name"`                      // 检测脚本名称
	Kind        string    `json:"kind"`                             // 检查的资源类型
	CheckDesc   string    `json:"check_desc"`                       // 检查脚本内容描述
	Cluster     string    `json:"cluster"`                          // 检查集群
	Namespace   string    `json:"namespace"`                        // 资源命名空间
	Name        string    `json:"name"`                             // 资源名称
	CreatedAt   time.Time `json:"created_at,omitempty" gorm:"<-:create"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`  // Automatically managed by GORM for update time
	ScheduleID  *uint     `json:"schedule_id,omitempty"` // 关联的定时任务ID
}

// List 返回符合条件的 InspectionCheckEvent 列表及总数
func (c *InspectionCheckEvent) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*InspectionCheckEvent, int64, error) {
	return dao.GenericQuery(params, c, queryFuncs...)
}

// Save 保存或更新 InspectionCheckEvent 实例
func (c *InspectionCheckEvent) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, c, queryFuncs...)
}

// Delete 根据指定 ID 删除 InspectionCheckEvent 实例
func (c *InspectionCheckEvent) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, c, utils.ToInt64Slice(ids), queryFuncs...)
}

// GetOne 获取单个 InspectionCheckEvent 实例
func (c *InspectionCheckEvent) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*InspectionCheckEvent, error) {
	return dao.GenericGetOne(params, c, queryFuncs...)
}

// BatchSave 批量保存 InspectionCheckEvent 实例
func (c *InspectionCheckEvent) BatchSave(params *dao.Params, events []*InspectionCheckEvent, batchSize int, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericBatchSave(params, events, batchSize, queryFuncs...)
}
