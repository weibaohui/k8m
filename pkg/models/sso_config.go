package models

import (
	"time"
)

// SSOConfig SSO配置表
type SSOConfig struct {
	ID                 uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	Name               string    `gorm:"not null" json:"name,omitempty"`                    // 配置名称
	Type               string    `gorm:"default:oidc" json:"type,omitempty"`                // 配置类型
	ClientID           string    `gorm:"type:text;not null" json:"client_id,omitempty"`     // OAuth2客户端ID
	ClientSecret       string    `gorm:"type:text;not null" json:"client_secret,omitempty"` // OAuth2客户端密钥
	RedirectURL        string    `gorm:"type:text;not null" json:"redirect_url,omitempty"`  // 回调URL
	Issuer             string    `gorm:"type:text;not null" json:"issuer,omitempty"`        // 认证服务器地址
	Enabled            bool      `gorm:"default:false" json:"enabled,omitempty"`            // 是否启用SSO
	AutoDiscovery      bool      `gorm:"default:true" json:"auto_discovery,omitempty"`      // 是否启用自动发现
	PreferUserNameKeys string    `gorm:"type:text;" json:"prefer_user_name_keys,omitempty"` // 用户自定义获取用户名的字段顺序，适用于如果用户名字段不在默认字段中情况
	AuthURL            string    `gorm:"type:text;" json:"auth_url,omitempty"`              // 认证URL
	TokenURL           string    `gorm:"type:text;" json:"token_url,omitempty"`             // Token获取URL
	UserInfoURL        string    `gorm:"type:text;" json:"user_info_url,omitempty"`         // 用户信息获取URL
	Scopes             []string  `gorm:"type:json" json:"scopes,omitempty"`                 // OAuth2授权范围
	CreatedAt          time.Time `json:"created_at,omitempty"`                              // 创建时间
	UpdatedAt          time.Time `json:"updated_at,omitempty"`                              // 更新时间
	CreatedBy          string    `json:"created_by,omitempty"`
}
