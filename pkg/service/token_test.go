package service

import (
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/weibaohui/k8m/pkg/flag"
	"github.com/weibaohui/k8m/pkg/models"
	"gorm.io/gorm"
)

// newTestTokenService 返回使用独立 sqlite 文件库的 TokenService，避免污染全局单例连接。
func newTestTokenService(t *testing.T) *tokenService {
	t.Helper()
	_ = flag.Init() // 确保 flag 包完成一次初始化（不影响本测试的DB注入）
	db, err := gorm.Open(sqlite.Open(t.TempDir()+"/token_test.db"), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开测试库失败: %v", err)
	}
	if err := db.AutoMigrate(&models.UserSession{}, &models.User{}); err != nil {
		t.Fatalf("迁移测试表失败: %v", err)
	}
	return &tokenService{testDB: db}
}

// TestTokenIssueAndValidate 签发后立即可校验，且返回正确用户名。
func TestTokenIssueAndValidate(t *testing.T) {
	svc := newTestTokenService(t)
	token, err := svc.Issue("alice", "password", "127.0.0.1", time.Hour)
	if err != nil {
		t.Fatalf("签发失败: %v", err)
	}
	if !strings.HasPrefix(token, "k8m_") {
		t.Errorf("token 应带 k8m_ 前缀，实际: %s", token)
	}
	if len(token) < 64 {
		t.Errorf("token 熵不足: 长度=%d", len(token))
	}
	got, err := svc.Validate(token)
	if err != nil {
		t.Fatalf("校验失败: %v", err)
	}
	if got != "alice" {
		t.Errorf("用户名不匹配: got=%s want=alice", got)
	}
}

// TestTokenHashNotPlaintext 库中只存哈希，不落明文 token。
func TestTokenHashNotPlaintext(t *testing.T) {
	svc := newTestTokenService(t)
	token, err := svc.Issue("bob", "password", "", time.Hour)
	if err != nil {
		t.Fatalf("签发失败: %v", err)
	}
	var count int64
	if err := svc.db().Model(&models.UserSession{}).Where("token_hash = ?", token).Count(&count).Error; err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	if count != 0 {
		t.Error("明文 token 不应出现在数据库中")
	}
	var sessionCount int64
	if err := svc.db().Model(&models.UserSession{}).Where("username = ?", "bob").Count(&sessionCount).Error; err != nil {
		t.Fatalf("查询失败: %v", err)
	}
	if sessionCount != 1 {
		t.Errorf("应恰好存在1条会话记录，实际=%d", sessionCount)
	}
}

// TestTokenRevoke 撤销后校验立即失败。
func TestTokenRevoke(t *testing.T) {
	svc := newTestTokenService(t)
	token, _ := svc.Issue("carol", "password", "", time.Hour)
	if _, err := svc.Validate(token); err != nil {
		t.Fatalf("签发后校验应通过: %v", err)
	}
	if err := svc.Revoke(token); err != nil {
		t.Fatalf("撤销失败: %v", err)
	}
	if _, err := svc.Validate(token); err == nil {
		t.Error("撤销后校验应失败，实际通过")
	}
}

// TestTokenRevokeAllByUser 改密码/禁用场景：按用户全撤后，其名下所有会话全部失效。
func TestTokenRevokeAllByUser(t *testing.T) {
	svc := newTestTokenService(t)
	t1, _ := svc.Issue("dave", "password", "", time.Hour)
	t2, _ := svc.Issue("dave", "mcp_key", "", 10*365*24*time.Hour)
	if err := svc.RevokeAllByUser("dave"); err != nil {
		t.Fatalf("全撤失败: %v", err)
	}
	if _, err := svc.Validate(t1); err == nil {
		t.Error("全撤后 t1 校验应失败")
	}
	if _, err := svc.Validate(t2); err == nil {
		t.Error("全撤后 t2 校验应失败")
	}
}

// TestTokenNegativeDuration 负有效期应被拒绝签发（防呆：不静默转正）。
func TestTokenNegativeDuration(t *testing.T) {
	svc := newTestTokenService(t)
	if _, err := svc.Issue("frank", "password", "", -time.Minute); err == nil {
		t.Error("负有效期签发应报错，实际成功")
	}
}

// TestTokenInvalid 捏造的 token 不可通过校验。
func TestTokenInvalid(t *testing.T) {
	svc := newTestTokenService(t)
	if _, err := svc.Validate("k8m_fake_token_value"); err == nil {
		t.Error("捏造 token 校验应失败")
	}
	if _, err := svc.Validate(""); err == nil {
		t.Error("空 token 校验应失败")
	}
}
