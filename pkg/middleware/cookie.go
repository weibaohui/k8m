package middleware

import (
	"net/http"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/flag"
	"github.com/weibaohui/k8m/pkg/service"
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

		// 检查请求路径是否需要跳过集群检测
		if path == "/" ||
			path == "/favicon.ico" ||
			path == "/auth/login" ||
			strings.HasPrefix(path, "/assets/") ||
			strings.HasPrefix(path, "/params/") || // 配置参数
			strings.HasPrefix(path, "/mgm/") || // 个人中心
			strings.HasPrefix(path, "/admin/") || // 管理后台
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
		// 注意，这里的role，可能是一个数组，需要处理一下
		role := claims[constants.JwtUserRole].(string)
		cst := claims[constants.JwtClusters].(string)
		roles := strings.Split(role, ",")
		csts := strings.Split(cst, ",")
		// klog.V(6).Infof("username=%s", username)

		var clusterID string
		allClusters := service.ClusterService().AllClusters()
		// 检查是否存在名为 "selectedCluster" 的 Cookie
		sc, err := c.Cookie("selectedCluster")
		if err != nil {
			// 不存在cookie

			if slices.Contains(roles, constants.RolePlatformAdmin) {
				// 平台管理员，选一个
				clusterID = service.ClusterService().FirstClusterID()
			} else {
				// 其他的只能选有权限中的某一个
				if len(csts) > 0 {
					clusterID = csts[0]
				}

			}

			if clusterID == "" {
				// 没有集群，说明不是管理员，也没有权限，那么就不要访问了
				c.JSON(512, gin.H{
					"msg": "无可用集群",
				})
				c.Abort()
				return
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

		// 如果sc为空，说明没有选择，不能直接使用
		if sc == "" {
			c.JSON(512, gin.H{
				"msg": "请选择集群",
			})
			c.Abort()
			return
		}

		// 如果设置了sc，但是集群未连接
		if !service.ClusterService().IsConnected(sc) {
			c.JSON(512, gin.H{
				"msg": "集群未连接，请先连接: " + sc,
			})
			c.Abort()
			return
		}

		// 如果不是平台管理员，检查是否有权限访问该集群
		if !slices.Contains(roles, constants.RolePlatformAdmin) && !slices.Contains(csts, sc) {
			c.JSON(512, gin.H{
				"msg": "无权限访问集群: " + sc,
			})
			c.Abort()
			return
		}
		// 继续处理下一个中间件或最终路由
		c.Next()
	}
}
