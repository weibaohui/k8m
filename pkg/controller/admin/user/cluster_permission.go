package user

import (
	"fmt"
	"strings"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/kom"
	"gorm.io/gorm"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

type AdminClusterPermission struct{}

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

}

// @Summary 获取指定集群指定角色的用户权限列表
// @Security BearerAuth
// @Param cluster path string true "集群ID(base64)"
// @Param role path string true "角色"
// @Success 200 {object} string
// @Router /admin/cluster_permissions/cluster/{cluster}/role/{role}/user/list [get]
func (a *AdminClusterPermission) ListClusterPermissions(c *gin.Context) {
	clusterBase64 := c.Param("cluster")
	role := c.Param("role")
	cluster, err := utils.DecodeBase64(clusterBase64)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	params := dao.BuildParams(c)
	m := &models.ClusterUserRole{}
	m.Cluster = cluster
	m.Role = role
	queryFuncs := genQueryFuncs(c, params)
	queryFuncs = append(queryFuncs, func(db *gorm.DB) *gorm.DB {
		return db.Where(m).Order("authorization_type desc ,username asc")
	})
	items, total, err := m.List(params, queryFuncs...)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonListWithTotal(c, total, items)
}

// @Summary 获取指定用户已获得授权的集群
// @Security BearerAuth
// @Param username path string true "用户名"
// @Success 200 {object} string
// @Router /admin/cluster_permissions/user/{username}/list [get]
func (a *AdminClusterPermission) ListClusterPermissionsByUserName(c *gin.Context) {
	username := c.Param("username")
	clusters, err := service.UserService().GetClusters(username)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonList(c, clusters)
}

// @Summary 获取指定集群下所有用户的权限角色列表
// @Security BearerAuth
// @Param cluster path string true "集群ID(base64)"
// @Success 200 {object} string
// @Router /admin/cluster_permissions/cluster/{cluster}/list [get]
func (a *AdminClusterPermission) ListClusterPermissionsByClusterID(c *gin.Context) {
	clusterBase64 := c.Param("cluster")
	cluster, err := utils.DecodeBase64(clusterBase64)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	params := dao.BuildParams(c)
	m := &models.ClusterUserRole{}
	m.Cluster = cluster
	items, total, err := m.List(params, func(db *gorm.DB) *gorm.DB {
		return db.Where(m).Order("authorization_type desc ,username asc")
	})
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonListWithTotal(c, total, items)
}

// @Summary 获取指定集群下所有命名空间名称
// @Security BearerAuth
// @Param cluster path string true "集群ID(base64)"
// @Success 200 {object} string
// @Router /admin/cluster_permissions/cluster/{cluster}/ns/list [get]
func (a *AdminClusterPermission) ListClusterNamespaceListByClusterID(c *gin.Context) {
	ctx := amis.GetContextWithUser(c)
	clusterBase64 := c.Param("cluster")
	cluster, err := utils.DecodeBase64(clusterBase64)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	var list []*v1.Namespace
	err = kom.Cluster(cluster).WithContext(ctx).Resource(&v1.Namespace{}).List(&list).Error

	if err != nil {
		amis.WriteJsonData(c, gin.H{
			"options": make([]map[string]string, 0),
		})
		return
	}
	var names []map[string]string
	for _, n := range list {
		names = append(names, map[string]string{
			"label": n.Name,
			"value": n.Name,
		})
	}
	slice.SortBy(names, func(a, b map[string]string) bool {
		return a["label"] < b["label"]
	})
	amis.WriteJsonData(c, gin.H{
		"options": names,
	})

}

// @Summary 批量为指定集群添加用户角色权限
// @Security BearerAuth
// @Param cluster path string true "集群ID(base64)"
// @Param role path string true "角色"
// @Param authorization_type path string true "授权类型"
// @Success 200 {object} string
// @Router /admin/cluster_permissions/cluster/{cluster}/role/{role}/{authorization_type}/save [post]
func (a *AdminClusterPermission) SaveClusterPermission(c *gin.Context) {
	clusterBase64 := c.Param("cluster")
	role := c.Param("role")
	authorizationType := c.Param("authorization_type")
	cluster, err := utils.DecodeBase64(clusterBase64)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	// {"users":"lisi,no2fa,test"}
	type requestBody struct {
		Users string `json:"users"`
	}
	var userList requestBody

	err = c.ShouldBindJSON(&userList)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	if userList.Users == "" {
		amis.WriteJsonError(c, fmt.Errorf("用户列表不能为空"))
		return
	}

	params := dao.BuildParams(c)

	// 默认授权类型为用户
	if authorizationType == "" {
		authorizationType = "user"
	}
	for _, username := range strings.Split(userList.Users, ",") {
		var m models.ClusterUserRole
		m.Cluster = cluster
		m.Role = role
		m.Username = username
		m.AuthorizationType = constants.ClusterAuthorizationType(authorizationType)
		one, err := m.GetOne(params, func(db *gorm.DB) *gorm.DB {
			return db.Where(m)
		})

		if err != nil || one == nil {
			// 不在用户权限条目，则添加
			err := m.Save(params)
			if err != nil {
				klog.V(6).Infof("新增用户权限失败: %s", err.Error())
				continue
			}
		}

		// 如果存在该集群下的用户条目，则跳过，不做处理
	}
	service.UserService().ClearCacheByKey("cluster")
	amis.WriteJsonOK(c)
}

// @Summary 删除集群权限
// @Security BearerAuth
// @Param ids path string true "权限ID，多个用逗号分隔"
// @Success 200 {object} string
// @Router /admin/cluster_permissions/delete/{ids} [post]
func (a *AdminClusterPermission) DeleteClusterPermission(c *gin.Context) {
	ids := c.Param("ids")

	params := dao.BuildParams(c)
	params.UserName = "" // 避免增加CreatedBy字段,因为用户是管理员创建的，所以不需要CreatedBy字段
	m := &models.ClusterUserRole{}

	err := m.Delete(params, ids)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	service.UserService().ClearCacheByKey("cluster")

	amis.WriteJsonOK(c)
}

// @Summary 更新指定集群用户角色的命名空间字段
// @Security BearerAuth
// @Param id path int true "权限ID"
// @Success 200 {object} string
// @Router /admin/cluster_permissions/update_namespaces/{id} [post]
func (a *AdminClusterPermission) UpdateNamespaces(c *gin.Context) {
	id := c.Param("id")
	type requestBody struct {
		Namespaces string `json:"namespaces"`
	}
	var nsList requestBody

	err := c.ShouldBindJSON(&nsList)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	params := dao.BuildParams(c)
	m := &models.ClusterUserRole{}
	m.ID = utils.ToUInt(id)
	m.Namespaces = nsList.Namespaces
	err = m.Save(params, func(db *gorm.DB) *gorm.DB {
		return db.Select("namespaces")
	})
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	service.UserService().ClearCacheByKey("cluster")

	amis.WriteJsonOK(c)
}

// @Summary 更新指定集群用户角色的黑名单命名空间字段
// @Security BearerAuth
// @Param id path int true "权限ID"
// @Success 200 {object} string
// @Router /admin/cluster_permissions/update_blacklist_namespaces/{id} [post]
func (a *AdminClusterPermission) UpdateBlacklistNamespaces(c *gin.Context) {
	id := c.Param("id")
	type requestBody struct {
		BlacklistNamespaces string `json:"blacklist_namespaces"`
	}
	var nsList requestBody

	err := c.ShouldBindJSON(&nsList)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	params := dao.BuildParams(c)
	m := &models.ClusterUserRole{}
	m.ID = utils.ToUInt(id)
	m.BlacklistNamespaces = nsList.BlacklistNamespaces
	err = m.Save(params, func(db *gorm.DB) *gorm.DB {
		return db.Select("blacklist_namespaces")
	})
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	service.UserService().ClearCacheByKey("cluster")

	amis.WriteJsonOK(c)
}
