package middleware

import (
	"encoding/base64"
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

func decodeUrlSafeBase64(s string) ([]byte, error) {
	// 补等号
	if m := len(s) % 4; m != 0 {
		s += strings.Repeat("=", 4-m)
	}
	return base64.URLEncoding.DecodeString(s)
}
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
			path == "/auth/login" ||
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
		clusterIDByte, _ := decodeUrlSafeBase64(clusterBase64)
		clusterID := string(clusterIDByte)

		if clusterID == "" {
			c.JSON(512, gin.H{
				"msg": "请先选择集群",
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
