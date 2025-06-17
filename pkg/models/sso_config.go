package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

// SSOConfig SSO配置表
type SSOConfig struct {
	ID                 uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	Name               string    `json:"name,omitempty"`                                    // 配置名称
	Type               string    `gorm:"default:oidc" json:"type,omitempty"`                // 配置类型
	ClientID           string    `gorm:"type:text;" json:"client_id,omitempty"`             // OAuth2客户端ID
	ClientSecret       string    `gorm:"type:text;" json:"client_secret,omitempty"`         // OAuth2客户端密钥
	Issuer             string    `gorm:"type:text;" json:"issuer,omitempty"`                // 认证服务器地址
	Enabled            bool      `gorm:"default:false" json:"enabled,omitempty"`            // 是否启用SSO
	PreferUserNameKeys string    `gorm:"type:text;" json:"prefer_user_name_keys,omitempty"` // 用户自定义获取用户名的字段顺序，适用于如果用户名字段不在默认字段中情况
	Scopes             string    `gorm:"type:text;" json:"scopes,omitempty"`                // 授权范围
	CreatedAt          time.Time `json:"created_at,omitempty"`                              // 创建时间
	UpdatedAt          time.Time `json:"updated_at,omitempty"`                              // 更新时间
}

// List 列出所有记录
func (s *SSOConfig) List(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) ([]*SSOConfig, int64, error) {
	return dao.GenericQuery(params, s, queryFuncs...)
}

// Save 保存记录
func (s *SSOConfig) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, s, queryFuncs...)
}

// Delete 删除记录
func (s *SSOConfig) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, s, utils.ToInt64Slice(ids), queryFuncs...)
}

// GetOne 获取单条记录
func (s *SSOConfig) GetOne(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) (*SSOConfig, error) {
	return dao.GenericGetOne(params, s, queryFuncs...)
}
