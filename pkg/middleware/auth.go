package middleware

import (
	"context"
	"net/http"
	"slices"
	"strings"

	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/response"
	"github.com/weibaohui/k8m/pkg/service"
)

// extractToken 从请求头或查询参数中提取凭证。
// 兼容 "Bearer xxx" 头与 ?token=xxx 查询参数（SSE/下载等场景使用）。
func extractToken(r *http.Request) string {
	token := r.Header.Get("Authorization")
	if token == "" {
		token = r.URL.Query().Get("token")
	}
	token = strings.TrimPrefix(token, "Bearer ")
	return strings.TrimSpace(token)
}

// AuthMiddleware 登录校验（Opaque Token + 服务端会话）。
// 取代原 JWT 校验：凭证为随机 token，校验走 TokenService（库中哈希比对），
// 不存在可伪造的签名密钥（issue #482 修复）。
func AuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := response.New(w, r)
			// 获取请求路径
			path := r.URL.Path
			// 检查请求路径是否需要跳过登录检测
			if path == "/" ||
				path == "/favicon.ico" ||
				path == "/ping" ||
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

			token := extractToken(r)
			username, err := service.TokenService().Validate(token)
			if err != nil {
				c.JSON(http.StatusUnauthorized, response.H{"message": err.Error()})
				return
			}

			// 设置信息传递，后面才能从ctx中获取到用户信息
			ctx := context.WithValue(r.Context(), constants.JwtUserName, username)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// PlatformAuthMiddleware 平台管理员角色校验（Opaque Token + 服务端会话 + DB角色检查）。
func PlatformAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := response.New(w, r)
			token := extractToken(r)
			username, err := service.TokenService().Validate(token)
			if err != nil {
				c.JSON(http.StatusUnauthorized, response.H{"message": err.Error()})
				return
			}
			if username == "" {
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
			ctx := context.WithValue(r.Context(), constants.JwtUserName, username)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
