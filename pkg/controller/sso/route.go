package sso

import (
	"github.com/gin-gonic/gin"
)

type AuthController struct{}

func RegisterAuthRoutes(auth *gin.RouterGroup) {
	ctrl := &AuthController{}
	auth.GET("/sso/config", ctrl.GetSSOConfig)
	auth.GET("/oidc/:name/sso", ctrl.GetAuthCodeURL)
	auth.GET("/oidc/:name/callback", ctrl.HandleCallback)
	auth.GET("/ldap/config", ctrl.GetLdapEnabled)
}
