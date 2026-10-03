package models

import (
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"gorm.io/gorm"
)

// UserSession 登录会话（Opaque Token）记录表。
// 设计目标：替代 JWT 作为登录态凭证，修复默认 JWT Secret 可伪造平台管理员 Token 的安全问题（issue #482）。
// 凭证本体只下发给用户，库中仅存 SHA-256 哈希；删除/撤销记录即该凭证立即失效。
type UserSession struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	Username       string    `gorm:"size:255;index:idx_user_session_username;not null" json:"username,omitempty"` // 登录用户名
	TokenHash      string    `gorm:"size:64;uniqueIndex:idx_user_session_token_hash;not null" json:"-"`           // token 的 SHA-256 十六进制哈希，明文不落库
	SourceType     string    `gorm:"size:32" json:"source_type,omitempty"`                                        // 签发来源：password / ldap / sso / temp_admin
	IPAddress      string    `gorm:"size:64" json:"ip_address,omitempty"`                                         // 登录来源 IP
	ExpiresAt      time.Time `gorm:"index:idx_user_session_expires" json:"expires_at,omitempty"`                  // 过期时间
	LastSeenAt     time.Time `json:"last_seen_at,omitempty"`                                                      // 最近一次使用时间
	CreatedAt      time.Time `json:"created_at,omitempty" gorm:"<-:create"`
	UpdatedAt      time.Time `json:"updated_at,omitempty"`
	RevokedAt      *time.Time `json:"revoked_at,omitempty"` // 非空即已撤销
}

func (c *UserSession) Save(params *dao.Params, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericSave(params, c, queryFuncs...)
}

func (c *UserSession) Delete(params *dao.Params, ids string, queryFuncs ...func(*gorm.DB) *gorm.DB) error {
	return dao.GenericDelete(params, c, utils.ToInt64Slice(ids), queryFuncs...)
}
