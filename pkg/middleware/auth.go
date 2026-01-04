package middleware

import (
	"context"
	"net/http"
	"slices"
	"strings"

	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/flag"
	"github.com/weibaohui/k8m/pkg/response"
	"github.com/weibaohui/k8m/pkg/service"
)

// AuthMiddleware 登录校验 - Gin到Chi迁移
func AuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := response.New(w, r)
			// 获取请求路径
			path := r.URL.Path
			// 检查请求路径是否需要跳过登录检测
			if path == "/" ||
				path == "/favicon.ico" ||
				path == "/healthz" ||
				strings.HasPrefix(path, "/monacoeditorwork/") ||
				strings.HasPrefix(path, "/swagger/") ||
				strings.HasPrefix(path, "/debug/") ||
				strings.HasPrefix(path, "/health/") ||
				strings.HasPrefix(path, "/mcp/") ||
				strings.HasPrefix(path, "/auth/") ||
				strings.HasPrefix(path, "/assets/") ||
				strings.HasPrefix(path, "/public/") {
				next.ServeHTTP(w, r)
				return

			}

			cfg := flag.Init()
			claims, err := utils.GetJWTClaims(c, cfg.JwtTokenSecret)
			if err != nil {
				c.JSON(http.StatusUnauthorized, response.H{"message": err.Error()})
				return
			}

			// 设置信息传递，后面才能从ctx中获取到用户信息
			ctx := context.WithValue(r.Context(), constants.JwtUserName, claims[constants.JwtUserName])
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// PlatformAuthMiddleware 平台管理员角色校验 - Gin到Chi迁移
func PlatformAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := response.New(w, r)
			cfg := flag.Init()
			claims, err := utils.GetJWTClaims(c, cfg.JwtTokenSecret)
			if err != nil {
				c.JSON(http.StatusUnauthorized, response.H{"message": err.Error()})
				return
			}

			username, ok := claims[constants.JwtUserName].(string)
			if !ok || username == "" {
				c.JSON(http.StatusUnauthorized, response.H{"error": "无效的用户名"})
				return
			}
			roles, err := service.UserService().GetRolesByUserName(username)
			if err != nil {
				c.JSON(http.StatusInternalServerError, response.H{"error": "角色查询失败"})
				return
			}
			// 权限检查
			if !slices.Contains(roles, constants.RolePlatformAdmin) {
				c.JSON(http.StatusForbidden, response.H{"error": "平台管理员权限校验失败"})
				return
			}

			// 设置信息传递，后面才能从ctx中获取到用户信息
			ctx := context.WithValue(r.Context(), constants.JwtUserName, claims[constants.JwtUserName])
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
