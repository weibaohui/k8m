package amis

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/service"
)

func GetSelectedCluster(c *gin.Context) string {
	selectedCluster, _ := c.Cookie("selectedCluster")
	if selectedCluster == "" {
		selectedCluster = service.ClusterService().FirstClusterID()
	}
	return selectedCluster
}

// GetLoginUser 获取当前登录用户名及其角色
func GetLoginUser(c *gin.Context) (string, string) {
	user := c.GetString(constants.JwtUserName)
	role := c.GetString(constants.JwtUserRole)
	return user, role
}

func GetContextWithUser(c *gin.Context) *context.Context {
	user, role := GetLoginUser(c)
	ctx := context.WithValue(c.Request.Context(), constants.JwtUserName, user)
	ctx = context.WithValue(ctx, constants.JwtUserRole, role)
	return &ctx
}
