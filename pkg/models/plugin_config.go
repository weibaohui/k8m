package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

// PluginConfig 插件状态持久化模型
// 用于记录插件的配置状态（已发现/已安装/已启用/已禁用），启动时根据该表动态应用
type PluginConfig struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	Name        string    `gorm:"uniqueIndex;not null" json:"name,omitempty"`        // 插件名称，唯一
	Status      string    `gorm:"type:varchar(32);not null" json:"status,omitempty"` // 状态：discovered/installed/enabled/disabled
	Version     string    `gorm:"type:varchar(64)" json:"version,omitempty"`         // 当前数据库记录的插件版本
	Description string    `json:"description,omitempty"`                             // 描述信息
	CreatedAt   time.Time `json:"created_at,omitempty" gorm:"<-:create"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
	CreatedBy   string    `json:"created_by,omitempty"`
}

// List 列出所有记录
// 支持动态查询与分页
func (p *PluginConfig) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*PluginConfig, int64, error) {
	return dao.GenericQuery(params, p, queryFuncs...)
}

// Save 保存记录
// 直接保存当前记录
func (p *PluginConfig) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, p, queryFuncs...)
}

// SaveByName 根据名称进行保存或更新
// 若名称已存在则更新该记录，否则创建新记录
func (p *PluginConfig) SaveByName(params *dao.Params) error {
	var existing PluginConfig
	err := dao.DB().Model(&PluginConfig{}).Where("name = ?", p.Name).First(&existing).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return err
	}
	if err == nil {
		p.ID = existing.ID
	}
	return dao.GenericSave(params, p)
}

// Delete 删除记录
// 支持批量删除
func (p *PluginConfig) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, p, utils.ToInt64Slice(ids), queryFuncs...)
}

// GetOne 获取单条记录
// 支持自定义查询函数
func (p *PluginConfig) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*PluginConfig, error) {
	return dao.GenericGetOne(params, p, queryFuncs...)
}
