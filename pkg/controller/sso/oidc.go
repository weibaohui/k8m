package sso

import (
	"fmt"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/models"
	"golang.org/x/oauth2"
)

type Config struct {
	Issuer       string   // OIDC Issuer 地址
	ClientID     string   // 应用注册的 Client ID
	ClientSecret string   // 应用注册的 Secret
	RedirectURL  string   // 登录回调地址
	Scopes       []string // eg: ["openid", "email", "profile"]
}

type Client struct {
	OAuth2Config *oauth2.Config
	Provider     *oidc.Provider
	Verifier     *oidc.IDTokenVerifier
	DBConfig     *models.SSOConfig
}

// NewOIDCClient  创建一个 OIDC 客户端
func NewOIDCClient(c *gin.Context, cfg *models.SSOConfig) (*Client, error) {
	ctx := c.Request.Context()
	// 1. 探测 issuer 的元信息
	provider, err := oidc.NewProvider(ctx, cfg.Issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to query oidc provider: %w", err)
	}

	// 2. 构建 OIDC ID Token 验证器
	verifier := provider.Verifier(&oidc.Config{
		ClientID: cfg.ClientID,
	})

	// 3. 构建 OAuth2 配置
	// 从上下文中获取gin.Context
	var redirectURL string
	// 动态构建回调URL
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	host := c.Request.Host
	redirectURL = fmt.Sprintf("%s://%s/auth/%s/%s/callback", scheme, host, cfg.Type, cfg.Name)

	oauth2Config := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  redirectURL,
		Endpoint:     provider.Endpoint(),                                                               // 自动使用 /.well-known 配置的接口
		Scopes:       append([]string{"openid", "email", "profile"}, strings.Split(cfg.Scopes, ",")...), // openid 是必须的
	}

	return &Client{
		OAuth2Config: oauth2Config,
		Provider:     provider,
		Verifier:     verifier,
		DBConfig:     cfg,
	}, nil
}
