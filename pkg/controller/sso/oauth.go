package sso

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/service"
	"k8s.io/klog/v2"
)

// HandleCallback 处理OAuth2回调
func HandleCallback(c *gin.Context) {
	ctx := c.Request.Context()
	client, err := NewOIDCClient(ctx, Config{
		Issuer:       "http://localhost:5556",
		ClientID:     "example-app",
		ClientSecret: "example-app-secret",
		RedirectURL:  "http://localhost:3000/auth/callback",
		Scopes:       []string{"email", "profile"},
	})

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
	if err := idToken.Claims(&claims); err != nil {
		c.String(http.StatusInternalServerError, "Parse claims error: %v", err)
		return
	}

	// c.JSON(http.StatusOK, gin.H{
	// 	"access_token": token.AccessToken,
	// 	"user_info":    userInfo,
	// })
	username := claims["email"].(string)
	service.UserService().CheckAndCreateUser(username, "sso")
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
func GetAuthCodeURL(c *gin.Context) {
	client, err := NewOIDCClient(c.Request.Context(), Config{
		Issuer:       "http://localhost:5556",
		ClientID:     "example-app",
		ClientSecret: "example-app-secret",
		RedirectURL:  "http://localhost:3000/auth/callback",
		Scopes:       []string{"email", "profile"},
	})
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	state := utils.RandNLengthString(8)
	url := client.OAuth2Config.AuthCodeURL(state)
	klog.Infof("url: %s", url)
	c.Redirect(http.StatusFound, url)
}
