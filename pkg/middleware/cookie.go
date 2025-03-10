package middleware

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/flag"
	"github.com/weibaohui/k8m/pkg/service"
	"k8s.io/klog/v2"
	"slices"
)

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

		cfg := flag.Init()
		var clusterID string
		allClusters := service.ClusterService().AllClusters()
		// 检查是否存在名为 "selectedCluster" 的 Cookie
		sc, err := c.Cookie("selectedCluster")
		if err != nil {
			// 不存在cookie
			clusterID = service.ClusterService().FirstClusterID()

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
		// InCluster模式下，只有一个集群，那么就直接用InCluster
		if cfg.InCluster && len(allClusters) == 1 && sc != "InCluster" {
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
		// 非集群内模式下，但是用了InCluster，肯定不对，需要纠正过来
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
		if sc != "" && !service.ClusterService().IsConnected(sc) {
			// 前端跳转到集群选择页面
			// 所以要排除集群页面的路径
			path := c.Request.URL.Path
			klog.V(6).Infof("c.Request.URL.Path=%s", path)
			if !(path == "/" ||
				path == "/favicon.ico" ||
				path == "/auth/login" ||
				strings.HasPrefix(path, "/assets/") ||
				strings.HasPrefix(path, "/public/") ||
				strings.Contains(path, "/cluster/file/option_list") ||
				strings.Contains(path, "/cluster/scan") ||
				strings.Contains(path, "/cluster/all") ||
				strings.Contains(path, "/cluster/kubeconfig/save") ||
				strings.Contains(path, "/cluster/kubeconfig/remove") ||
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
