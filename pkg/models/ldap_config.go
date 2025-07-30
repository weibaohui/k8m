package models

import (
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
	"time"
)

// LDAPConfig 用于存储LDAP配置
// bind_password 字段加密存储
// gorm.Model 包含ID、CreatedAt、UpdatedAt、DeletedAt

type LDAPConfig struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	Name            string    `gorm:"size:50;not null" json:"name"`            // 配置名称
	Host            string    `gorm:"size:100;not null" json:"host"`           // 服务器地址
	Port            int       `gorm:"not null" json:"port"`                    // 端口
	BindDN          string    `gorm:"size:100;not null" json:"bind_dn"`        // 管理员DN
	BindPassword    string    `gorm:"not null" json:"bind_password,omitempty"` // 管理员密码（加密存储）
	BaseDN          string    `gorm:"size:100;not null" json:"base_dn"`        // 基础DN
	UserFilter      string    `gorm:"size:255" json:"user_filter"`             // 用户过滤器
	LOGIN2AUTHCLOSE bool      `gorm:"default:true" json:"login2_auth_close"`   // 登录后开启认证
	DefaultGroup    string    `gorm:"size:50" json:"default_group"`            // 默认用户组
	Enabled         bool      `gorm:"default:true" json:"enabled"`             // 启用状态
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// List 列出所有记录
func (l *LDAPConfig) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*LDAPConfig, int64, error) {
	return dao.GenericQuery(params, l, queryFuncs...)
}

// Save 保存记录
func (l *LDAPConfig) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, l, queryFuncs...)
}

// Delete 删除记录
func (l *LDAPConfig) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, l, utils.ToInt64Slice(ids), queryFuncs...)
}

// GetOne 获取单条记录
func (l *LDAPConfig) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*LDAPConfig, error) {
	return dao.GenericGetOne(params, l, queryFuncs...)
}
