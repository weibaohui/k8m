package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"k8s.io/klog/v2"
)

// CustomRecovery 是自定义的 Recovery 中间件
func CustomRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 打印错误日志（可选）
				klog.Errorf("捕获到 panic: %v\n", err)
				// 返回友好的错误信息
				amis.WriteJsonError(c, fmt.Errorf("服务器内部错误，请稍后再试。"))
				c.Abort()
			}
		}()
		c.Next() // 继续执行下一个中间件或处理函数
	}
}
