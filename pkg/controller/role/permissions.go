package role

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/service"
)

type BindingRequest struct {
	TargetType string `json:"target_type" binding:"required"` // 'user' 或 'group'
	TargetID   string `json:"target_id" binding:"required"`   // 用户ID或组ID
	ClusterID  string `json:"cluster_id" binding:"required"`
	Namespace  string `json:"namespace" binding:"required"`
	RoleID     string `json:"role_id" binding:"required"`
}

type ListBindingsRequest struct {
	Page     int    `form:"page" json:"page"`
	PageSize int    `form:"pageSize" json:"pageSize"`
	Keyword  string `form:"keyword" json:"keyword"`
}

type ListBindingsResponse struct {
	Total int64       `json:"total"`
	Items interface{} `json:"items"`
}

// CreateBinding 创建新的权限绑定
func CreateBinding(c *gin.Context) {
	var req BindingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 获取当前用户名
	username, exists := c.Get(constants.JwtUserName)
	if !exists {
		amis.WriteJsonError(c, errors.New("未找到用户信息"))
		return
	}

	permService := service.PermissionService()
	bindingID, err := permService.CreatePermissionBinding(
		req.TargetType,
		req.TargetID,
		req.ClusterID,
		req.Namespace,
		req.RoleID,
		username.(string),
	)

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonData(c, gin.H{
		"binding_id": bindingID,
	})
}

// DeleteBinding 删除权限绑定
func DeleteBinding(c *gin.Context) {
	bindingID := c.Param("binding_id")
	if bindingID == "" {
		amis.WriteJsonError(c, errors.New("缺少绑定ID"))
		return
	}

	permService := service.PermissionService()
	if err := permService.DeletePermissionBinding(bindingID); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonOK(c)
}

// ListBindings 列出权限绑定
func ListBindings(c *gin.Context) {
	var req ListBindingsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	permService := service.PermissionService()
	permissions, total, err := permService.ListPermissionBindings(c.Request.Context(), req.Page, req.PageSize, req.Keyword)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonData(c, ListBindingsResponse{
		Total: total,
		Items: permissions,
	})
}

// CheckUserPermission 检查当前用户是否拥有特定权限
func CheckUserPermission(c *gin.Context) {
	// 获取当前用户名
	username, exists := c.Get(constants.JwtUserName)
	if !exists {
		amis.WriteJsonError(c, errors.New("未找到用户信息"))
		return
	}

	// 获取请求参数
	clusterID := c.Query("cluster_id")
	namespace := c.Query("namespace")
	operation := c.Query("operation")

	if clusterID == "" || operation == "" {
		amis.WriteJsonError(c, errors.New("缺少集群ID或操作类型"))
		return
	}

	permService := service.PermissionService()

	// 获取用户ID
	userID, err := permService.GetUserID(username.(string))
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 检查权限
	allowed, err := permService.CheckPermission(userID, clusterID, namespace, operation)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonData(c, gin.H{
		"allowed": allowed,
	})
}
