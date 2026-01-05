package middleware

import (
	"fmt"
	"net/http"

	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/response"
	"k8s.io/klog/v2"
)

// CustomRecovery 是自定义的 Recovery 中间件 - Gin到Chi迁移
func CustomRecovery() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// 打印错误日志（可选）
					klog.Errorf("捕获到 panic: %v\n", err)
					// 返回友好的错误信息
					c := response.New(w, r)
					amis.WriteJsonError(c, fmt.Errorf("服务器内部错误，请稍后再试。"))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
