package models

import (
	"gorm.io/gorm"
)

// SSOConfig SSO配置表
type SSOConfig struct {
	gorm.Model
	Name               string   `gorm:"not null" json:"name"`                    // 配置名称
	ClientID           string   `gorm:"type:text;not null" json:"client_id"`     // OAuth2客户端ID
	ClientSecret       string   `gorm:"type:text;not null" json:"client_secret"` // OAuth2客户端密钥
	RedirectURL        string   `gorm:"type:text;not null" json:"redirect_url"`  // 回调URL
	Issuer             string   `gorm:"type:text;not null" json:"issuer"`        // 认证服务器地址
	Enabled            bool     `gorm:"default:false" json:"enabled"`            // 是否启用SSO
	AutoDiscovery      bool     `gorm:"default:true" json:"auto_discovery"`      // 是否启用自动发现
	PreferUserNameKeys string   `gorm:"type:text;" json:"prefer_user_name_keys"` // 用户自定义获取用户名的字段顺序，适用于如果用户名字段不在默认字段中情况
	AuthURL            string   `gorm:"type:text;" json:"auth_url"`              // 认证URL
	TokenURL           string   `gorm:"type:text;" json:"token_url"`             // Token获取URL
	UserInfoURL        string   `gorm:"type:text;" json:"user_info_url"`         // 用户信息获取URL
	Scopes             []string `gorm:"type:json" json:"scopes"`                 // OAuth2授权范围
}
