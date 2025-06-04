package middleware

import (
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/flag"
)

// AuthMiddleware 登录校验
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取请求路径
		path := c.Request.URL.Path
		// 检查请求路径是否需要跳过登录检测
		if path == "/" ||
			path == "/favicon.ico" ||
			path == "/healthz" ||
			strings.HasPrefix(path, "/swagger/") ||
			strings.HasPrefix(path, "/debug/") ||
			strings.HasPrefix(path, "/mcp/") ||
			strings.HasPrefix(path, "/auth/") ||
			strings.HasPrefix(path, "/assets/") ||
			strings.HasPrefix(path, "/public/") {
			c.Next()
			return

		}

		cfg := flag.Init()
		claims, err := utils.GetJWTClaims(c, cfg.JwtTokenSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
			c.Abort()

			return
		}

		// 设置信息传递，后面才能从ctx中获取到用户信息
		c.Set(constants.JwtUserName, claims[constants.JwtUserName])
		c.Set(constants.JwtUserRole, claims[constants.JwtUserRole])
		c.Set(constants.JwtClusters, claims[constants.JwtClusters])
		// 判断 claims[constants.JwtClusterUserRoles]的类型，应该是[]models.ClusterUserRole
		// 为什么会出现string？
		c.Set(constants.JwtClusterUserRoles, claims[constants.JwtClusterUserRoles])
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
		role := claims[constants.JwtUserRole].(string)

		// 权限检查
		roles := strings.Split(role, ",")
		if !slices.Contains(roles, constants.RolePlatformAdmin) {
			c.JSON(http.StatusForbidden, gin.H{"error": "平台管理员权限校验失败"})
			c.Abort()
			return
		}

		// 设置信息传递，后面才能从ctx中获取到用户信息
		c.Set(constants.JwtUserName, claims[constants.JwtUserName])
		c.Set(constants.JwtUserRole, claims[constants.JwtUserRole])
		c.Set(constants.JwtClusters, claims[constants.JwtClusters])
		c.Set(constants.JwtClusterUserRoles, claims[constants.JwtClusterUserRoles])
		c.Next()
	}
}
