package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/service"
)

func EnsureSelectedClusterMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否存在名为 "selectedCluster" 的 Cookie
		_, err := c.Cookie("selectedCluster")
		if err != nil {
			clusterID := service.ClusterService().FirstClusterID()
			if clusterID == "" {
				c.Next()
			} else {
				// 如果 Cookie 不存在，写入一个默认值
				c.SetCookie(
					"selectedCluster",           // Cookie 名称
					clusterID,                   // Cookie 默认值
					int(24*time.Hour.Seconds()), // 有效期（秒），这里是 1 天
					"/",                         // Cookie 路径
					"",                          // 域名，默认当前域
					false,                       // 是否仅 HTTPS
					false,                       // 是否 HttpOnly
				)
			}

		}

		// 继续处理下一个中间件或最终路由
		c.Next()
	}
}
