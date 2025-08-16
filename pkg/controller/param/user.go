package param

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/service"
)

// UserRole 获取当前用户的Role信息
// @Summary 获取用户角色信息
// @Description 获取当前登录用户的角色及默认集群
// @Security BearerAuth
// @Success 200 {object} string
// @Router /params/user/role [get]
func (pc *Controller) UserRole(c *gin.Context) {
	user, role := amis.GetLoginUser(c)
	clusters := amis.GetLoginUserClusters(c)
	var cluster string
	if len(clusters) == 1 {
		cluster = clusters[0]
	}
	if cluster == "" {
		if amis.IsCurrentUserPlatformAdmin(c) {
			cluster = service.ClusterService().FirstClusterID()
		}
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
	amis.WriteJsonData(c, gin.H{
		"role":     role,
		"cluster":  cluster,
		"menuData": menuData,
	})
}
