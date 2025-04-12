package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

// ConditionReverse 用于记录需要反转解释的K8s状态指标
type ConditionReverse struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	Name        string    `gorm:"type:text" json:"name,omitempty"`        // 指标名称，使用包含方式查找。如Pressure、Unavailable等
	Enabled     bool      `json:"enabled,omitempty"`                      // 指标描述
	Description string    `gorm:"type:text" json:"description,omitempty"` // 指标描述
	CreatedAt   time.Time `json:"created_at,omitempty"`                   // 创建时间
	UpdatedAt   time.Time `json:"updated_at,omitempty"`                   // 更新时间
}

// List 列出所有记录
func (c *ConditionReverse) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*ConditionReverse, int64, error) {
	return dao.GenericQuery(params, c, queryFuncs...)
}

// Save 保存记录
func (c *ConditionReverse) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, c, queryFuncs...)
}

// Delete 删除记录
func (c *ConditionReverse) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, c, utils.ToInt64Slice(ids), queryFuncs...)
}

// GetOne 获取单条记录
func (c *ConditionReverse) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*ConditionReverse, error) {
	return dao.GenericGetOne(params, c, queryFuncs...)
}
