package middleware

import (
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
)

// 定义方法元数据
type Controller struct{}

// 定义方法，并用 struct tag 标记访问权限
func (c Controller) AdminOnly(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "Admin action executed"})
}

func (c Controller) UserOrAdmin(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "User or Admin action executed"})
}

// 反射解析 struct tag 获取权限要求
func PermissionMiddleware(handler interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetHeader("X-Role")
		if role == "" {
			role = "guest"
		}

		// 通过反射获取方法名
		handlerValue := reflect.ValueOf(handler)
		handlerType := handlerValue.Type()

		// 获取 struct tag
		requiredRole := handlerType.Name()

		// 权限检查
		if (requiredRole == "AdminOnly" && role != "admin") ||
			(requiredRole == "UserOrAdmin" && role != "admin" && role != "user") {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access Denied for your role"})
			c.Abort()
			return
		}

		// 继续执行请求处理
		handlerValue.Call([]reflect.Value{reflect.ValueOf(c)})
	}
}
