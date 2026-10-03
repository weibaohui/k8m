package service

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/models"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

// tokenService 基于 Opaque Token 的会话服务。
// 替代 JWT：凭证为 256 位随机数，库中仅存哈希；校验=哈希比对+过期+撤销检查，
// 天然支持撤销（删除/禁用用户/改密码即失效），不存在可伪造的签名密钥（issue #482 根治）。
type tokenService struct {
	// testDB 测试注入用。为空时使用全局 dao.DB()。
	testDB *gorm.DB
}

var localTokenService = &tokenService{}

func TokenService() *tokenService {
	return localTokenService
}

// db 返回操作句柄：测试注入优先，否则全局连接。
func (t *tokenService) db() *gorm.DB {
	if t.testDB != nil {
		return t.testDB
	}
	return dao.DB()
}

const (
	// tokenCacheTTL 会话校验缓存时长。撤销后最多延迟该时长生效，换取每请求免查库。
	tokenCacheTTL = 30 * time.Second
	// DefaultSessionDuration 默认浏览器会话有效期
	DefaultSessionDuration = 24 * time.Hour
)

// Issue 为用户签发新会话凭证，返回明文 token（仅此一次可见），库中落哈希。
// duration 必须为正数（未来有效期），传负值视为无效请求直接报错，不静默转换为默认值。
func (t *tokenService) Issue(username, sourceType, ip string, duration time.Duration) (string, error) {
	if username == "" {
		return "", errors.New("username cannot be empty")
	}
	if duration <= 0 {
		return "", errors.New("duration 必须为正数")
	}
	token, err := utils.GenerateSecureToken()
	if err != nil {
		return "", err
	}
	now := time.Now()
	session := &models.UserSession{
		Username:   username,
		TokenHash:  t.hash(token),
		SourceType: sourceType,
		IPAddress:  ip,
		ExpiresAt:  now.Add(duration),
		LastSeenAt: now,
	}
	if err := t.db().Create(session).Error; err != nil {
		return "", fmt.Errorf("保存会话失败: %w", err)
	}
	return token, nil
}

// AdoptLegacyToken 把升级前已签发给用户的凭证（旧 JWT 格式的 API 密钥/MCP 密钥）登记为服务端会话，
// 使其在切换到 Opaque Token 会话后继续可用。返回登记条数。
//
// 安全性：只登记"确实存在于密钥表中的凭证"，攻击者用公开默认密钥伪造的 JWT 不在表中，登记后依然无效，
// 因此兼容迁移不引入伪造风险。
// 幂等：已存在同哈希且未撤销的会话则跳过，可重复调用。
func (t *tokenService) AdoptLegacyToken(username, token, sourceType string, expiresAt time.Time) (bool, error) {
	if username == "" || token == "" {
		return false, nil
	}
	hash := t.hash(token)
	var count int64
	if err := t.db().Model(&models.UserSession{}).Where("token_hash = ?", hash).Count(&count).Error; err != nil {
		return false, err
	}
	if count > 0 {
		return false, nil
	}
	if expiresAt.IsZero() || !expiresAt.After(time.Now()) {
		// 无过期时间或已过期：按默认会话时长补一个有效期，避免登记出立即失效的会话
		expiresAt = time.Now().Add(DefaultSessionDuration)
	}
	session := &models.UserSession{
		Username:   username,
		TokenHash:  hash,
		SourceType: sourceType,
		ExpiresAt:  expiresAt,
		LastSeenAt: time.Now(),
	}
	if err := t.db().Create(session).Error; err != nil {
		return false, fmt.Errorf("登记存量凭证失败: %w", err)
	}
	return true, nil
}

// Validate 校验 token，返回用户名；无效/过期/已撤销返回错误。
// 结果按 token 哈希缓存 30 秒，撤销操作会主动清缓存。
func (t *tokenService) Validate(token string) (string, error) {
	if token == "" {
		return "", errors.New("未提供 Token")
	}
	hash := t.hash(token)
	cacheKey := "token:session:" + hash
	username, err := utils.GetOrSetCache(CacheService().CacheInstance(), cacheKey, tokenCacheTTL, func() (string, error) {
		session := &models.UserSession{}
		if err := t.db().Where("token_hash = ?", hash).First(session).Error; err != nil {
			return "", errors.New("Token 无效")
		}
		if session.RevokedAt != nil {
			return "", errors.New("Token 已被撤销")
		}
		if time.Now().After(session.ExpiresAt) {
			return "", errors.New("Token 已过期")
		}
		// 最近使用时间随缓存刷新同步更新（每凭证每 30 秒至多一次）
		if err := t.db().Model(&models.UserSession{}).Where("token_hash = ?", hash).Update("last_seen_at", time.Now()).Error; err != nil {
			klog.V(6).Infof("更新会话最近使用时间失败: %v", err)
		}
		return session.Username, nil
	})
	if err != nil {
		return "", err
	}
	return username, nil
}

// Revoke 撤销单个 token（用户登出场景可调用；前端清除 localStorage 即可完成登出，此方法供 API 管理）。
func (t *tokenService) Revoke(token string) error {
	hash := t.hash(token)
	now := time.Now()
	if err := t.db().Model(&models.UserSession{}).Where("token_hash = ?", hash).Update("revoked_at", now).Error; err != nil {
		return err
	}
	t.clearCache(hash)
	return nil
}

// RevokeAllByUser 撤销某用户的全部有效会话（禁用用户、删除用户、修改密码时调用）。
func (t *tokenService) RevokeAllByUser(username string) error {
	sessions := []*models.UserSession{}
	if err := t.db().Where("username = ? and revoked_at is null", username).Find(&sessions).Error; err != nil {
		return err
	}
	now := time.Now()
	if err := t.db().Model(&models.UserSession{}).Where("username = ? and revoked_at is null", username).Update("revoked_at", now).Error; err != nil {
		return err
	}
	for _, s := range sessions {
		t.clearCache(s.TokenHash)
	}
	if len(sessions) > 0 {
		klog.V(4).Infof("已撤销用户[%s]的%d个会话", username, len(sessions))
	}
	return nil
}

// RevokeAllByUserByIDs 按用户ID列表撤销其全部有效会话（删除用户场景：先取用户名再撤销）。
func (t *tokenService) RevokeAllByUserByIDs(ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	users := []*models.User{}
	if err := t.db().Where("id in ?", ids).Find(&users).Error; err != nil {
		return err
	}
	for _, u := range users {
		if err := t.RevokeAllByUser(u.Username); err != nil {
			return err
		}
	}
	return nil
}

// CleanupExpired 删除已过期超过 grace 的会话记录，由启动后的定时任务周期调用。
func (t *tokenService) CleanupExpired() (int64, error) {
	db := t.db().Where("expires_at < ?", time.Now().Add(-24*time.Hour))
	var count int64
	if err := db.Model(&models.UserSession{}).Count(&count).Error; err != nil {
		return 0, err
	}
	if count == 0 {
		return 0, nil
	}
	return count, db.Delete(&models.UserSession{}).Error
}

func (t *tokenService) hash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (t *tokenService) clearCache(hash string) {
	utils.ClearCacheByKey(CacheService().CacheInstance(), "token:session:"+hash)
}
