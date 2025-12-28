package param

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/service"
)

// UserRole 获取当前用户的Role信息
// @Summary 获取用户角色信息
// @Description 获取当前登录用户的角色及默认集群
// @Security BearerAuth
// @Success 200 {object} string
// @Router /params/user/role [get]
func (pc *Controller) UserRole(c *gin.Context) {
	user := amis.GetLoginUser(c)

	// 如果是平台管理员,可以看到所有菜单
	if service.UserService().IsUserPlatformAdmin(user) {
		amis.WriteJsonData(c, gin.H{
			"roles":     []string{constants.RolePlatformAdmin},
			"cluster":   "",
			"groups":    []string{},
			"menu_data": []any{},
		})
		return
	}

	groupNames, err := service.UserService().GetGroupNames(user)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	menuData, err := service.UserService().GetGroupMenuData(groupNames)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	roles, err := service.UserService().GetRolesByUserName(user)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	cluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, gin.H{
		"role":      roles,
		"cluster":   cluster,
		"groups":    groupNames,
		"menu_data": menuData,
	})
}
