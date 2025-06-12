package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

// InspectionLuaScript 表示一条 Lua 脚本的元数据及内容
// 包含脚本名称、描述、分组、版本、类型和脚本内容等信息
// 用于存储和管理自定义 Lua 脚本
type InspectionLuaScript struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	Name        string    `gorm:"primaryKey;size:128" json:"name"` // 脚本名称，主键
	Description string    `json:"description"`                     // 脚本描述
	Group       string    `json:"group"`                           // 分组
	Version     string    `json:"version"`                         // 版本
	Kind        string    `json:"kind"`                            // 类型
	Script      string    `gorm:"type:text" json:"script"`         // 脚本内容
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"` // Automatically managed by GORM for update time
	CreatedBy   string    `json:"created_by,omitempty"` // 创建者

}

// List 返回符合条件的 InspectionLuaScript 列表及总数
func (c *InspectionLuaScript) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*InspectionLuaScript, int64, error) {
	return dao.GenericQuery(params, c, queryFuncs...)
}

// Save 保存或更新 InspectionLuaScript 实例
func (c *InspectionLuaScript) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, c, queryFuncs...)
}

// Delete 根据指定 ID 删除 InspectionLuaScript 实例
func (c *InspectionLuaScript) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, c, utils.ToInt64Slice(ids), queryFuncs...)
}

// GetOne 获取单个 InspectionLuaScript 实例
func (c *InspectionLuaScript) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*InspectionLuaScript, error) {
	return dao.GenericGetOne(params, c, queryFuncs...)
}
