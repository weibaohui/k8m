package models

import (
	"fmt"
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"

	"gorm.io/gorm"
)

// InspectionRecord 用于记录每次巡检的发起和执行信息，包括定时和手动触发
// 关联 InspectionSchedule（可选），并可关联多个脚本执行结果
// @author: AI
// @date: 2024-05-18
// @desc: 巡检执行记录表

type InspectionRecord struct {
	ID           uint       `gorm:"primaryKey;autoIncrement" json:"id,omitempty"` // 主键
	ScheduleID   *uint      `json:"schedule_id,omitempty"`                        // 关联的定时任务ID，可为空（手动触发时为空）
	ScheduleName string     `json:"schedule_name,omitempty"`                      // 巡检任务名称快照
	Cluster      string     `json:"cluster"`                                      // 巡检目标集群
	TriggerType  string     `json:"trigger_type"`                                 // 触发类型（manual/cron）
	Status       string     `json:"status"`                                       // 执行状态（pending/running/success/failed）
	StartTime    time.Time  `json:"start_time"`
	EndTime      *time.Time `json:"end_time,omitempty"`
	CreatedAt    time.Time  `json:"created_at,omitempty" gorm:"<-:create"`
	UpdatedAt    time.Time  `json:"updated_at,omitempty"` // Automatically managed by GORM for update time
	ErrorCount   int        `json:"error_count"`
	AISummary    string     `json:"ai_summary,omitempty"` // AI生成的巡检总结
}

// InspectionScriptResult 记录每个巡检脚本的执行结果，关联到 InspectionRecord
// @author: AI
// @date: 2024-05-18
// @desc: 巡检脚本执行结果表

type InspectionScriptResult struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	RecordID   uint      `json:"record_id"`   // 关联的巡检执行记录ID
	ScriptName string    `json:"script_name"` // 脚本名称
	ScriptKind string    `json:"script_kind"` // 脚本资源类型
	ScriptDesc string    `json:"script_desc"` // 脚本描述
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	StdOutput  string    `json:"std_output"`          // 脚本标准输出
	ErrorMsg   string    `json:"error_msg,omitempty"` // 错误信息
	CreatedAt  time.Time `json:"created_at,omitempty" gorm:"<-:create"`
	UpdatedAt  time.Time `json:"updated_at,omitempty"`  // Automatically managed by GORM for update time
	Cluster    string    `json:"cluster"`               // 目标集群
	ScheduleID *uint     `json:"schedule_id,omitempty"` // 巡检计划ID
}

// List 返回符合条件的 InspectionRecord 列表及总数
func (c *InspectionRecord) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*InspectionRecord, int64, error) {
	return dao.GenericQuery(params, c, queryFuncs...)
}

// Save 保存或更新 InspectionRecord 实例
func (c *InspectionRecord) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, c, queryFuncs...)
}

// Delete 删除指定ID的 InspectionRecord
func (c *InspectionRecord) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, c, utils.ToInt64Slice(ids), queryFuncs...)
}

// GetOne 获取单个 InspectionRecord
func (c *InspectionRecord) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*InspectionRecord, error) {
	return dao.GenericGetOne(params, c, queryFuncs...)
}

// GetAISummaryById 获取 InspectionRecord 的AISummary
func (c *InspectionRecord) GetAISummaryById(recordID uint) (string, error) {
	record := &InspectionRecord{}
	record, err := record.GetOne(nil, func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", recordID)
	})
	if err != nil {
		return "", fmt.Errorf("未找到对应的巡检记录: %d", recordID)
	}
	return record.AISummary, nil
}

// List 返回符合条件的 InspectionScriptResult 列表及总数
func (c *InspectionScriptResult) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*InspectionScriptResult, int64, error) {
	return dao.GenericQuery(params, c, queryFuncs...)
}

// Save 保存或更新 InspectionScriptResult 实例
func (c *InspectionScriptResult) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, c, queryFuncs...)
}

// Delete 删除指定ID的 InspectionScriptResult
func (c *InspectionScriptResult) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, c, utils.ToInt64Slice(ids), queryFuncs...)
}

// GetOne 获取单个 InspectionScriptResult
func (c *InspectionScriptResult) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*InspectionScriptResult, error) {
	return dao.GenericGetOne(params, c, queryFuncs...)
}

// BatchSave 批量保存 InspectionCheckEvent 实例
func (c *InspectionScriptResult) BatchSave(params *dao.Params, events []*InspectionScriptResult, batchSize int, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericBatchSave(params, events, batchSize, queryFuncs...)
}
