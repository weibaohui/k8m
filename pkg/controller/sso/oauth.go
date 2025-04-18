package sso

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/service"
	"golang.org/x/oauth2"
)

var (
	clientID     = "example-app"
	clientSecret = "example-app-secret"
	redirectURL  = "http://localhost:3618/callback"
	issuer       = "http://localhost:5556"

	oauthConfig = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  issuer + "/auth",
			TokenURL: issuer + "/token",
		},
		Scopes: []string{"openid", "email", "profile"},
	}
)

// HandleCallback 处理OAuth2回调
func HandleCallback(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.String(http.StatusBadRequest, "No code in request")
		return
	}

	// 用授权码换取 token
	token, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		c.String(http.StatusInternalServerError, "Token exchange error: %v", err)
		return
	}

	// 使用 token 获取 userinfo
	client := oauthConfig.Client(context.Background(), token)
	resp, err := client.Get(issuer + "/userinfo")
	if err != nil {
		c.String(http.StatusInternalServerError, "Userinfo fetch error: %v", err)
		return
	}
	defer resp.Body.Close()

	var userInfo map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		c.String(http.StatusInternalServerError, "Decode error: %v", err)
		return
	}

	// c.JSON(http.StatusOK, gin.H{
	// 	"access_token": token.AccessToken,
	// 	"user_info":    userInfo,
	// })
	username := userInfo["email"].(string)
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
	// 跳转到 Dex 的 /auth
	url := oauthConfig.AuthCodeURL(utils.RandNLengthString(8), oauth2.AccessTypeOffline)
	c.Redirect(http.StatusFound, url)
}
