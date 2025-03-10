package amis

import (
	"context"
	"strings"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/models"
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

	roles := strings.Split(role, ",")
	// 优先检查平台管理员
	if slice.Contain(roles, models.RolePlatformAdmin) {
		role = models.RolePlatformAdmin
	} else if slice.Contain(roles, models.RoleClusterAdmin) {
		// 其次检查集群管理员
		role = models.RoleClusterAdmin
	} else {
		// 默认设为只读
		role = models.RoleClusterReadonly
	}

	return user, role
}

func GetContextWithUser(c *gin.Context) context.Context {
	user, role := GetLoginUser(c)
	ctx := context.WithValue(c.Request.Context(), constants.JwtUserName, user)
	ctx = context.WithValue(ctx, constants.JwtUserRole, role)
	return ctx
}
