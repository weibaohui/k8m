package models

import (
	"time"

	"github.com/robfig/cron/v3"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

// InspectionSchedule 用于描述定时巡检任务的元数据，包括任务名称、描述、目标集群、cron表达式等
// 该结构体可用于存储和管理定时巡检任务的相关信息
// 字段涵盖任务名称、描述、目标集群、cron表达式、启用状态、AI总结功能配置、创建人和创建时间
// 支持配置AI总结功能的开关和自定义提示词模板
// 可结合数据库或配置管理进行持久化
type InspectionSchedule struct {
	ID                  uint         `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	Name                string       `json:"name"`                                // 巡检任务名称
	Description         string       `json:"description"`                         // 巡检任务描述
	Clusters            string       `json:"clusters"`                            // 目标集群列表
	Webhooks            string       `json:"webhooks"`                            // webhook列表
	WebhookNames        string       `json:"webhook_names"`                       // webhook 名称列表
	Cron                string       `json:"cron"`                                // cron表达式，定时周期
	ScriptCodes         string       `gorm:"type:text" json:"script_codes"`       // 每个脚本唯一标识码
	Enabled             bool         `json:"enabled"`                             // 是否启用该任务
	AIEnabled           bool         `json:"ai_enabled"`                          // 是否启用AI总结功能
	AIPromptTemplate    string       `gorm:"type:text" json:"ai_prompt_template"` // AI总结提示词模板
	CronRunID           cron.EntryID `json:"cron_run_id"`                         // cron 运行ID，可用于删除
	LastRunTime         *time.Time   `json:"last_run_time"`                       // 上次运行时间
	ErrorCount          int          `json:"error_count"`                         // 错误次数
	SkipZeroFailedCount bool         `json:"skip_zero_failed_count"`              // 是否跳过0失败的条目
	CreatedAt           time.Time    `json:"created_at,omitempty" gorm:"<-:create"`
	UpdatedAt           time.Time    `json:"updated_at,omitempty"` // Automatically managed by GORM for update time
}

// List 返回符合条件的 InspectionSchedule 列表及总数
func (c *InspectionSchedule) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*InspectionSchedule, int64, error) {
	return dao.GenericQuery(params, c, queryFuncs...)
}

// Save 保存或更新 InspectionSchedule 实例
func (c *InspectionSchedule) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, c, queryFuncs...)
}

// Delete 根据指定 ID 删除 InspectionSchedule 实例
func (c *InspectionSchedule) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, c, utils.ToInt64Slice(ids), queryFuncs...)
}

// GetOne 获取单个 InspectionSchedule 实例
func (c *InspectionSchedule) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*InspectionSchedule, error) {
	return dao.GenericGetOne(params, c, queryFuncs...)
}

// CheckSkipZeroFailedCount 检查巡检计划是否配置了跳过0失败的条目
func (c *InspectionSchedule) CheckSkipZeroFailedCount(id *uint) bool {
	if id == nil {
		klog.V(6).Infof("无法检查跳过0失败配置: id == nil")
		return false
	}

	schedule, err := dao.GenericGetOne(nil, c, func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id)
	})
	if err != nil {
		klog.V(6).Infof("获取巡检计划id=%d失败: %v", *id, err)
		return false
	}
	skip := schedule.SkipZeroFailedCount
	klog.V(4).Infof("巡检计划id=%d配置了跳过0失败的条目=%v", *id, skip)
	return skip
}
