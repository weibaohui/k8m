package models

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"k8s.io/klog/v2"
)

// legacyMcpKeyDuration 存量 MCP 密钥（旧版为 10 年期 JWT）迁移时补的有效期，与原签发时长保持一致。
const legacyMcpKeyDuration = 10 * 365 * 24 * time.Hour

// AdoptLegacyKeysAsSessions 把升级前已签发的密钥凭证登记为服务端会话，使其在切换到
// Opaque Token 校验后继续可用，用户无需重新生成密钥。
//
// 背景：旧版 API 密钥（api_keys.key）与 MCP 密钥（mcp_keys.jwt）存放的是 JWT 字符串，
// 校验依赖签名验证；改造后校验统一走 user_sessions 表按哈希比对，旧凭证若无会话记录会全部失效。
//
// 覆盖范围：
//   - api_keys.key  → source_type=api_key，过期时间取 api_keys.expires_at
//   - mcp_keys.jwt  → source_type=mcp_key，过期时间按签发时间 + 10 年推算（该表未存过期字段）
//
// 安全性：只登记密钥表中确实存在的凭证。用公开默认密钥伪造的 JWT 不存在于表中，
// 登记后依然无法通过校验，因此兼容迁移不引入伪造风险。
// 幂等：已存在同哈希会话的凭证会被跳过，每次启动重复执行无副作用。
func AdoptLegacyKeysAsSessions() {
	db := dao.DB()
	if db == nil {
		return
	}

	migrated := 0

	// API 密钥
	if db.Migrator().HasTable("api_keys") {
		type apiKeyRow struct {
			Username  string
			Key       string
			ExpiresAt time.Time
		}
		rows := []apiKeyRow{}
		if err := db.Table("api_keys").Select("username, key, expires_at").Find(&rows).Error; err != nil {
			klog.Errorf("读取存量API密钥失败: %v", err)
		}
		for _, r := range rows {
			expiresAt := r.ExpiresAt
			if expiresAt.IsZero() {
				expiresAt = time.Now().Add(24 * time.Hour)
			}
			if ok := adoptTokenAsSession(r.Username, r.Key, "api_key", expiresAt); ok {
				migrated++
			}
		}
	}

	// MCP 密钥
	if db.Migrator().HasTable("mcp_keys") {
		type mcpKeyRow struct {
			Username  string
			Jwt       string
			CreatedAt time.Time
		}
		rows := []mcpKeyRow{}
		if err := db.Table("mcp_keys").Select("username, jwt, created_at").Find(&rows).Error; err != nil {
			klog.Errorf("读取存量MCP密钥失败: %v", err)
		}
		for _, r := range rows {
			base := r.CreatedAt
			if base.IsZero() {
				base = time.Now()
			}
			if ok := adoptTokenAsSession(r.Username, r.Jwt, "mcp_key", base.Add(legacyMcpKeyDuration)); ok {
				migrated++
			}
		}
	}

	if migrated > 0 {
		klog.Infof("已将%d条存量密钥（API/MCP）登记为服务端会话，旧密钥继续可用", migrated)
	}
}

// adoptTokenAsSession 幂等地把一条既有凭证登记为会话，返回是否新增。
func adoptTokenAsSession(username, token, sourceType string, expiresAt time.Time) bool {
	if username == "" || token == "" {
		return false
	}
	db := dao.DB()
	hash := sha256Hex(token)
	var count int64
	if err := db.Model(&UserSession{}).Where("token_hash = ?", hash).Count(&count).Error; err != nil {
		klog.Errorf("检查存量凭证会话失败: %v", err)
		return false
	}
	if count > 0 {
		return false
	}
	if !expiresAt.After(time.Now()) {
		// 已过期或时间缺失：按默认会话时长补有效期，避免登记出立即失效的会话
		expiresAt = time.Now().Add(24 * time.Hour)
	}
	session := &UserSession{
		Username:   username,
		TokenHash:  hash,
		SourceType: sourceType,
		ExpiresAt:  expiresAt,
		LastSeenAt: time.Now(),
	}
	if err := db.Create(session).Error; err != nil {
		klog.Errorf("登记存量凭证失败: %v", err)
		return false
	}
	return true
}

// sha256Hex 返回字符串的 SHA-256 十六进制摘要，与 TokenService 的凭证哈希算法保持一致。
func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}
