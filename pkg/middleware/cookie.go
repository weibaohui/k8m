package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/flag"
	"github.com/weibaohui/k8m/pkg/service"
	"k8s.io/klog/v2"
)

func EnsureSelectedClusterMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {
		cfg := flag.Init()
		var clusterID string

		// 检查是否存在名为 "selectedCluster" 的 Cookie
		sc, err := c.Cookie("selectedCluster")
		if err != nil {
			// 不存在cookie
			if cfg.InCluster {
				clusterID = "InCluster"
			} else {
				// 从集群中选择一个
				clusterID = service.ClusterService().FirstClusterID()
			}

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
		if cfg.InCluster && sc != "InCluster" {
			// 集群内模式,但是当前cookie不是InCluster,那么给他纠正过来
			clusterID = "InCluster"
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
		if !cfg.InCluster && sc == "InCluster" {
			// 非集群内模式,但是当前cookie是InCluster,那么给他纠正过来
			clusterID = service.ClusterService().FirstClusterID()
			c.SetCookie(
				"selectedCluster",           // Cookie 名称
				clusterID,                   // Cookie 默认值
				int(24*time.Hour.Seconds()), // 有效期（秒），这里是 1 天
				"/",                         // Cookie 路径
				"",                          // 域名默认当前域
				false,                       // 是否仅 HTTPS
				false,                       // 是否 HttpOnly
			)
		}

		// 如果设置了sc，但是不能用
		if !service.ClusterService().IsConnected(sc) {
			// 前端跳转到集群选择页面
			// 所以要排除集群页面的路径
			path := c.Request.URL.Path
			klog.V(6).Infof("c.Request.URL.Path=%s", path)
			if !(path == "/" ||
				path == "/favicon.ico" ||
				strings.HasPrefix(path, "/assets/") ||
				strings.HasPrefix(path, "/public/") ||
				strings.Contains(path, "/cluster/file/option_list") ||
				strings.Contains(path, "/cluster/all") ||
				strings.Contains(path, "/cluster/reconnect") ||
				strings.Contains(path, "/cluster/setDefault")) {
				c.JSON(512, gin.H{
					"msg": sc,
				})
				c.Abort()
			}

		}
		// 继续处理下一个中间件或最终路由
		c.Next()
	}
}
