package param

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
)

// UserRole 获取当前用户的Role信息
func UserRole(c *gin.Context) {
	_, role := amis.GetLoginUser(c)
	amis.WriteJsonData(c, gin.H{
		"role": role,
	})
}
