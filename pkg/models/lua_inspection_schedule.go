package models

import (
	"time"

	"github.com/robfig/cron/v3"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

// InspectionSchedule 用于描述定时巡检任务的元数据，包括任务名称、描述、目标集群、cron表达式等
// 该结构体可用于存储和管理定时巡检任务的相关信息
// 字段涵盖任务名称、描述、目标集群、cron表达式、启用状态、创建人和创建时间
// 可结合数据库或配置管理进行持久化
type InspectionSchedule struct {
	ID          uint         `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	Name        string       `json:"name"`        // 巡检任务名称
	Description string       `json:"description"` // 巡检任务描述
	Clusters    string       `json:"clusters"`    // 目标集群列表
	Cron        string       `json:"cron"`        // cron表达式，定时周期
	Enabled     bool         `json:"enabled"`     // 是否启用该任务
	CreatedAt   time.Time    `json:"created_at,omitempty"`
	UpdatedAt   time.Time    `json:"updated_at,omitempty"` // Automatically managed by GORM for update time
	CronRunID   cron.EntryID `json:"cron_run_id"`          // cron 运行ID，可用于删除
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
