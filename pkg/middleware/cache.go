package middleware

import (
	"net/http"
	"path/filepath"
	"slices"
	"strconv"
	"time"
)

// SetCacheHeaders 设置静态文件缓存头 - Gin到Chi迁移
func SetCacheHeaders() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 获取请求路径
			path := r.URL.Path

			// 检查文件后缀，如果是静态文件则直接跳过
			ext := filepath.Ext(path)
			if ext == "" {
				// 无后缀，往往是后端请求
				next.ServeHTTP(w, r)
				return
			}
			if ext != "" {
				// 常见的静态文件后缀
				staticExts := []string{".js", ".css", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico", ".woff", ".woff2", ".ttf", ".eot", ".map"}
				if !slices.Contains(staticExts, ext) {
					// 静态文件请求，直接跳过集群检测
					next.ServeHTTP(w, r)
					return
				}
			}

			// 设置缓存时间为1小时（3600秒）
			maxAge := 3600 * 6
			w.Header().Set("Cache-Control", "public, max-age="+strconv.Itoa(maxAge))
			// 设置Expires头，为当前时间加上缓存时间
			expires := time.Now().Add(time.Second * time.Duration(maxAge)).Format(http.TimeFormat)
			w.Header().Set("Expires", expires)
			next.ServeHTTP(w, r)
		})
	}
}
