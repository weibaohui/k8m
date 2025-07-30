package sso

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/controller/admin/config"
)

type AuthController struct{}

func RegisterAuthRoutes(auth *gin.RouterGroup) {
	ctrl := &AuthController{}
	ldap := &config.LdapConfigController{}
	auth.GET("/sso/config", ctrl.GetSSOConfig)
	auth.GET("/oidc/:name/sso", ctrl.GetAuthCodeURL)
	auth.GET("/oidc/:name/callback", ctrl.HandleCallback)
	auth.GET("/ldap/config", ldap.GetLdapConfig)
}
