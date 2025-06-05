package amis

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/kom/kom"
)

func GetSelectedCluster(c *gin.Context) (string, error) {
	selectedCluster := c.GetString("cluster")
	if kom.Cluster(selectedCluster) == nil {
		return "", fmt.Errorf("cluster %s not found", selectedCluster)
	}
	return selectedCluster, nil
}

// GetLoginUserClusters 获取当前登录用户可访问集群列表
func GetLoginUserClusters(c *gin.Context) []string {
	cs := c.GetString(constants.JwtClusters)
	return strings.Split(cs, ",")
}

// GetLoginUser 获取当前登录用户名及其角色
func GetLoginUser(c *gin.Context) (string, string) {
	user := c.GetString(constants.JwtUserName)
	role := c.GetString(constants.JwtUserRole)

	roles := strings.Split(role, ",")
	role = constants.RoleGuest

	// 检查是否平台管理员
	if slice.Contain(roles, constants.RolePlatformAdmin) {
		role = constants.RolePlatformAdmin
	}
	return user, role
}

// GetLoginUserWithClusterRoles 获取当前登录用户名及其角色,已经授权的集群角色
// 返回值: 用户名, 角色, 集群角色列表
// Deprecated: 请使用UserService
func GetLoginUserWithClusterRoles(c *gin.Context) (string, string, []*models.ClusterUserRole) {
	user := c.GetString(constants.JwtUserName)
	role := c.GetString(constants.JwtUserRole)

	roles := strings.Split(role, ",")
	role = constants.RoleGuest

	// 检查是否平台管理员
	if slice.Contain(roles, constants.RolePlatformAdmin) {
		role = constants.RolePlatformAdmin
	}

	if value, exists := c.Get(constants.JwtClusterUserRoles); exists {
		switch v := value.(type) {
		case []*models.ClusterUserRole:
			return user, role, v
		case string:
			var clusterUserRoles []*models.ClusterUserRole
			if err := json.Unmarshal([]byte(v), &clusterUserRoles); err == nil {
				return user, role, clusterUserRoles
			}
		}

	}

	return user, role, nil

}

// IsCurrentUserPlatformAdmin 检测当前登录用户是否为平台管理员
func IsCurrentUserPlatformAdmin(c *gin.Context) bool {
	role := c.GetString(constants.JwtUserRole)
	roles := strings.Split(role, ",")
	return slice.Contain(roles, constants.RolePlatformAdmin)
}

func GetContextWithUser(c *gin.Context) context.Context {
	user, role, clusterRoles := GetLoginUserWithClusterRoles(c)
	ctx := context.WithValue(c.Request.Context(), constants.JwtUserName, user)
	ctx = context.WithValue(ctx, constants.JwtUserRole, role)
	ctx = context.WithValue(ctx, constants.JwtClusterUserRoles, clusterRoles)
	cst := ""
	for _, clusterRole := range clusterRoles {
		cst += clusterRole.Cluster + ","
	}
	ctx = context.WithValue(ctx, constants.JwtClusters, cst)

	return ctx
}
