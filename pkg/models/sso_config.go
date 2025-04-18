package models

import (
	"gorm.io/gorm"
)

// SSOConfig SSO配置表
type SSOConfig struct {
	gorm.Model
	ClientID      string   `gorm:"type:varchar(100);not null" json:"client_id"`     // OAuth2客户端ID
	ClientSecret  string   `gorm:"type:varchar(100);not null" json:"client_secret"` // OAuth2客户端密钥
	RedirectURL   string   `gorm:"type:varchar(200);not null" json:"redirect_url"`  // 回调URL
	Issuer        string   `gorm:"type:varchar(200);not null" json:"issuer"`        // 认证服务器地址
	Enabled       bool     `gorm:"default:false" json:"enabled"`                    // 是否启用SSO
	AutoDiscovery bool     `gorm:"default:true" json:"auto_discovery"`              // 是否启用自动发现
	AuthURL       string   `gorm:"type:varchar(200);not null" json:"auth_url"`      // 认证URL
	TokenURL      string   `gorm:"type:varchar(200);not null" json:"token_url"`     // Token获取URL
	UserInfoURL   string   `gorm:"type:varchar(200);not null" json:"user_info_url"` // 用户信息获取URL
	Scopes        []string `gorm:"type:json" json:"scopes"`                         // OAuth2授权范围
}
