package amis

import (
	"context"
	"encoding/json"
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
	role = models.RoleGuest

	// 检查是否平台管理员
	if slice.Contain(roles, models.RolePlatformAdmin) {
		role = models.RolePlatformAdmin
	}
	return user, role
}

// GetLoginUserWithClusterRoles 获取当前登录用户名及其角色,已经授权的集群角色
// 返回值: 用户名, 角色, 集群角色列表
func GetLoginUserWithClusterRoles(c *gin.Context) (string, string, []*models.ClusterUserRole) {
	user := c.GetString(constants.JwtUserName)
	role := c.GetString(constants.JwtUserRole)

	roles := strings.Split(role, ",")
	role = models.RoleGuest

	// 检查是否平台管理员
	if slice.Contain(roles, models.RolePlatformAdmin) {
		role = models.RolePlatformAdmin
	}

	clusterRoles := c.GetString(constants.JwtClusterUserRoles)
	var clusterUserRoles []*models.ClusterUserRole
	err := json.Unmarshal([]byte(clusterRoles), &clusterUserRoles)
	if err != nil {
		return user, role, nil
	}
	return user, role, clusterUserRoles
}

// IsCurrentUserPlatformAdmin 检测当前登录用户是否为平台管理员
func IsCurrentUserPlatformAdmin(c *gin.Context) bool {
	role := c.GetString(constants.JwtUserRole)
	roles := strings.Split(role, ",")
	return slice.Contain(roles, models.RolePlatformAdmin)
}

func GetContextWithUser(c *gin.Context) context.Context {
	user, role, clusterRoles := GetLoginUserWithClusterRoles(c)
	ctx := context.WithValue(c.Request.Context(), constants.JwtUserName, user)
	ctx = context.WithValue(ctx, constants.JwtUserRole, role)
	ctx = context.WithValue(ctx, constants.JwtClusterUserRoles, clusterRoles)
	return ctx
}
