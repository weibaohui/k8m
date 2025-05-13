package middleware

import (
	"net/http"
	"path/filepath"
	"slices"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func SetCacheHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {

		// 获取请求路径
		path := c.Request.URL.Path

		// 检查文件后缀，如果是静态文件则直接跳过
		ext := filepath.Ext(path)
		if ext == "" {
			// 无后缀，往往是后端请求
			c.Next()
			return
		}
		if ext != "" {
			// 常见的静态文件后缀
			staticExts := []string{".js", ".css", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico", ".woff", ".woff2", ".ttf", ".eot", ".map"}
			if !slices.Contains(staticExts, ext) {
				// 静态文件请求，直接跳过集群检测
				c.Next()
				return
			}
		}

		// 设置缓存时间为1小时（3600秒）
		maxAge := 3600 * 6
		c.Header("Cache-Control", "public, max-age="+strconv.Itoa(maxAge))
		// 设置Expires头，为当前时间加上缓存时间
		expires := time.Now().Add(time.Second * time.Duration(maxAge)).Format(http.TimeFormat)
		c.Header("Expires", expires)
		c.Next()
	}
}
