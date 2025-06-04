package middleware

import (
	"net/http"
	"path/filepath"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/flag"
	"github.com/weibaohui/k8m/pkg/service"
)

// EnsureSelectedClusterMiddleware 返回一个 Gin 中间件，用于强制校验请求是否已选择并有权限访问指定集群。
// 对于静态文件和部分白名单路径会直接跳过校验。其余请求将：
// 1. 校验 URL 中的集群参数是否存在且有效；
// 2. 校验用户 JWT 是否有效，并判断用户是否有访问该集群的权限（非平台管理员需在授权集群列表中）；
// 3. 校验目标集群是否已连接。
// 校验不通过时将中止请求并返回相应的错误信息。
func EnsureSelectedClusterMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {

		// 获取请求路径
		path := c.Request.URL.Path

		// 检查文件后缀，如果是静态文件则直接跳过
		ext := filepath.Ext(path)
		if ext != "" {
			// 常见的静态文件后缀
			staticExts := []string{".js", ".css", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico", ".woff", ".woff2", ".ttf", ".eot", ".map"}
			if slices.Contains(staticExts, ext) {
				// 静态文件请求，直接跳过集群检测
				c.Next()
				return
			}
		}

		// 检查请求路径是否需要跳过集群检测
		if path == "/" ||
			path == "/favicon.ico" ||
			path == "/healthz" ||
			strings.HasPrefix(path, "/swagger/") ||
			strings.HasPrefix(path, "/debug/") ||
			strings.HasPrefix(path, "/mcp/") ||
			strings.HasPrefix(path, "/auth/") ||
			strings.HasPrefix(path, "/assets/") ||
			strings.HasPrefix(path, "/ai/") || // ai 聊天不带cluster
			strings.HasPrefix(path, "/params/") || // 配置参数
			strings.HasPrefix(path, "/mgm/") || // 个人中心
			strings.HasPrefix(path, "/admin/") || // 管理后台
			strings.HasPrefix(path, "/public/") {
			c.Next()
			return

		}

		// 获取clusterID
		clusterBase64 := c.Param("cluster")
		clusterIDByte, _ := utils.UrlSafeBase64Decode(clusterBase64)
		clusterID := string(clusterIDByte)

		if clusterID == "" {
			c.JSON(512, gin.H{
				"msg": "未指定集群，请先切换集群",
			})
			c.Abort()
			return
		}

		cfg := flag.Init()
		claims, err := utils.GetJWTClaims(c, cfg.JwtTokenSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
			c.Abort()
			return
		}
		// 注意，这里的role，可能是一个数组，需要处理一下
		role := claims[constants.JwtUserRole].(string)
		cst := claims[constants.JwtClusters].(string)
		roles := strings.Split(role, ",")
		csts := strings.Split(cst, ",")
		// 如果不是平台管理员，检查是否有权限访问该集群
		if !slices.Contains(roles, constants.RolePlatformAdmin) && !slices.Contains(csts, clusterID) {
			c.JSON(512, gin.H{
				"msg": "无权限访问集群: " + clusterID,
			})
			c.Abort()
			return
		}

		// 如果设置了clusterID，但是集群未连接
		if !service.ClusterService().IsConnected(clusterID) {
			c.JSON(512, gin.H{
				"msg": "集群未连接，请先连接集群: " + clusterID,
			})
			c.Abort()
			return
		}
		c.Set("cluster", clusterID)

		// 继续处理下一个中间件或最终路由
		c.Next()
	}
}
