package user

import (
	"github.com/gin-gonic/gin"
)

// AdminClusterPermission 用于集群权限相关接口
// 路由注册函数
func RegisterClusterPermissionRoutes(admin *gin.RouterGroup) {
	ctrl := &AdminClusterPermission{}
	//  cluster_permissions 集群授权
	admin.GET("/cluster_permissions/cluster/:cluster/role/:role/user/list", ctrl.ListClusterPermissions)
	admin.GET("/cluster_permissions/user/:username/list", ctrl.ListClusterPermissionsByUserName)         // 列出指定用户拥有的集群权限
	admin.GET("/cluster_permissions/cluster/:cluster/list", ctrl.ListClusterPermissionsByClusterID)      // 列出指定集群下所有授权情况
	admin.GET("/cluster_permissions/cluster/:cluster/ns/list", ctrl.ListClusterNamespaceListByClusterID) // 列出指定集群下所有授权情况
	admin.POST("/cluster_permissions/cluster/:cluster/role/:role/:authorization_type/save", ctrl.SaveClusterPermission)
	admin.POST("/cluster_permissions/delete/:ids", ctrl.DeleteClusterPermission)
	admin.POST("/cluster_permissions/update_namespaces/:id", ctrl.UpdateNamespaces)
	admin.POST("/cluster_permissions/update_blacklist_namespaces/:id", ctrl.UpdateBlacklistNamespaces)

	// user 平台管理员可操作，管理用户
	admin.GET("/user/list", ctrl.List)
	admin.POST("/user/save", ctrl.Save)
	admin.POST("/user/delete/:ids", ctrl.Delete)
	admin.POST("/user/update_psw/:id", ctrl.UpdatePsw)
	admin.GET("/user/option_list", ctrl.UserOptionList)
	// 2FA 平台管理员可操作，管理用户
	admin.POST("/user/2fa/disable/:id", ctrl.Disable2FA)
	// user_group
	admin.GET("/user_group/list", ctrl.ListUserGroup)
	admin.POST("/user_group/save", ctrl.SaveUserGroup)
	admin.POST("/user_group/delete/:ids", ctrl.DeleteUserGroup)
	admin.GET("/user_group/option_list", ctrl.GroupOptionList)
}

type AdminClusterPermission struct{}
