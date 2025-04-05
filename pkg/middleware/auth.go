package middleware

import (
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/flag"
	"github.com/weibaohui/k8m/pkg/models"
)

// AuthMiddleware 登录校验
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := flag.Init()
		claims, err := utils.GetJWTClaims(c, cfg.JwtTokenSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
			c.Abort()
			return
		}

		c.Set(constants.JwtUserName, claims[constants.JwtUserName])
		c.Set(constants.JwtUserRole, claims[constants.JwtUserRole])
		c.Set(constants.JwtClusters, claims[constants.JwtClusters])
		c.Next()
	}
}

// PlatformAuthMiddleware 平台管理员角色校验
func PlatformAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := flag.Init()
		claims, err := utils.GetJWTClaims(c, cfg.JwtTokenSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
			c.Abort()
			return
		}
		username := claims[constants.JwtUserName].(string)
		role := claims[constants.JwtUserRole].(string)
		cst := claims[constants.JwtClusters].(string)

		// 权限检查
		roles := strings.Split(role, ",")
		if !slices.Contains(roles, models.RolePlatformAdmin) {
			c.JSON(http.StatusForbidden, gin.H{"error": "平台管理员权限校验失败"})
			c.Abort()
			return
		}

		c.Set(constants.JwtUserName, username)
		c.Set(constants.JwtUserRole, role)
		c.Set(constants.JwtClusters, cst)
		c.Next()
	}
}
