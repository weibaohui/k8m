package sso

import (
	"fmt"
	"github.com/weibaohui/k8m/pkg/flag"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/service"
	"k8s.io/klog/v2"
)

// GetSSOConfig 获取SSO列表，用于前端展示
func GetSSOConfig(c *gin.Context) {
	// 获取所有的SSO配置
	var ssoConfigs []models.SSOConfig
	err := dao.DB().Select([]string{"name", "type"}).Where("enabled == true").Find(&ssoConfigs).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, ssoConfigs)
}
func GetAuthCodeURL(c *gin.Context) {
	name := c.Param("name")
	klog.V(6).Infof("use sso name: %s", name)
	// 从配置文件中读取默认的OIDC客户端配置
	client, err := getDefaultOIDCClient(c, name)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	state := utils.RandNLengthString(8)
	url := client.OAuth2Config.AuthCodeURL(state)
	klog.Infof("url: %s", url)
	c.Redirect(http.StatusFound, url)
}

// HandleCallback 处理OAuth2回调
func HandleCallback(c *gin.Context) {
	name := c.Param("name")
	ctx := c.Request.Context()
	client, err := getDefaultOIDCClient(c, name)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	code := c.Query("code")
	oauth2Token, err := client.OAuth2Config.Exchange(ctx, code)
	if err != nil {
		c.String(http.StatusInternalServerError, "Token exchange error: %v", err)
		return
	}

	rawIDToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		c.String(http.StatusInternalServerError, "No id_token in token response")
		return
	}

	idToken, err := client.Verifier.Verify(ctx, rawIDToken)
	if err != nil {
		c.String(http.StatusInternalServerError, "Token verify error: %v", err)
		return
	}

	var claims map[string]interface{}
	if err = idToken.Claims(&claims); err != nil {
		c.String(http.StatusInternalServerError, "Parse claims error: %v", err)
		return
	}

	username := GetUsername(claims, strings.Split(client.DBConfig.PreferUserNameKeys, ","))
	_ = service.UserService().CheckAndCreateUser(username, name)
	userLoginToken, _ := service.UserService().GenerateJWTTokenByUserName(username, 24*time.Hour)

	// 返回 HTML + JS，用于写入 localStorage
	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
  <head><title>SSO Login Success</title></head>
  <body>
    <script>
      localStorage.setItem("token", %q);
      // 自动跳转回首页或 dashboard
      window.location.href = "/#/";
    </script>
    <p>登录成功，正在跳转...</p>
  </body>
</html>
`, userLoginToken)

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

// 获取默认OIDC客户端配置
func getDefaultOIDCClient(c *gin.Context, name string) (*Client, error) {
	// 通过name 获取配置
	var dbConfig *models.SSOConfig
	err := dao.DB().Where("name = ?", name).First(&dbConfig).Error
	if err != nil {
		return nil, err
	}

	return NewOIDCClient(c, dbConfig)

}
func GetUsername(claims map[string]interface{}, preferKeys []string) string {
	klog.V(6).Infof("claims: %v", claims)
	klog.V(6).Infof("preferKeys: %v", preferKeys)
	for _, key := range preferKeys {
		if val, ok := claims[key].(string); ok && val != "" {
			return val
		}
	}
	// 默认 fallback 顺序
	if v, ok := claims["preferred_username"].(string); ok && v != "" {
		return v
	}
	if v, ok := claims["email"].(string); ok && v != "" {
		return v
	}
	if v, ok := claims["name"].(string); ok && v != "" {
		return v
	}
	if v, ok := claims["sub"].(string); ok && v != "" {
		return v
	}
	return "unknown"
}

// 获取ldap开关状态
func GetLdapEnabled(c *gin.Context) {
	cfg := flag.Init()
	amis.WriteJsonData(c, gin.H{
		"enabled": cfg.LdapEnabled,
	})
}
