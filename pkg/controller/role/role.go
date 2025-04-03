package role

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/service"
)

type ListRoleRequest struct {
	Page     int    `form:"page" json:"page"`
	PageSize int    `form:"pageSize" json:"pageSize"`
	Keyword  string `form:"keyword" json:"keyword"`
}

type ListRoleResponse struct {
	Total int64       `json:"total"`
	Items interface{} `json:"items"`
}

type CreateRoleRequest struct {
	RoleName   string   `json:"roleName" binding:"required"`
	Operations []string `json:"operations" binding:"required"`
}

type UpdateRoleRequest struct {
	RoleName   string   `json:"roleName" binding:"required"`
	Operations []string `json:"operations" binding:"required"`
}

// List 获取角色列表
func List(c *gin.Context) {
	var req ListRoleRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 查询角色列表
	roles, total, err := service.RoleService().List(c, req.Page, req.PageSize, req.Keyword)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonData(c, ListRoleResponse{
		Total: total,
		Items: roles,
	})
}

// Create 创建角色
func Create(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	roleID, err := service.RoleService().Create(c, req.RoleName, req.Operations)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonData(c, gin.H{"role_id": roleID})
}

// Detail 获取角色详情
func Detail(c *gin.Context) {
	roleID := c.Param("role_id")
	role, err := service.RoleService().GetByID(c, roleID)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonData(c, role)
}

// Update 更新角色
func Update(c *gin.Context) {
	roleID := c.Param("role_id")
	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	err := service.RoleService().Update(c, roleID, req.RoleName, req.Operations)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonOK(c)
}

// Delete 删除角色
func Delete(c *gin.Context) {
	roleID := c.Param("role_id")
	err := service.RoleService().Delete(c, roleID)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonOK(c)
}
