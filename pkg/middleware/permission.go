package middleware

import (
	"net/http"
	"reflect"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/constants"
)

func RolePlatformOnly(handler interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, role := amis.GetLoginUser(c)
		if role == "" {
			role = "guest"
		}

		// 通过反射获取方法名
		handlerValue := reflect.ValueOf(handler)
		// handlerType := handlerValue.Type()

		// 获取 struct tag
		// requiredRole := handlerType.Name()

		// 权限检查
		roles := strings.Split(role, ",")
		if !slices.Contains(roles, constants.RolePlatformAdmin) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access Denied for your role"})
			c.Abort()
			return
		}

		// 继续执行请求处理
		handlerValue.Call([]reflect.Value{reflect.ValueOf(c)})
	}
}
