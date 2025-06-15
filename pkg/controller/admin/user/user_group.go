package user

import (
	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/service"
	"gorm.io/gorm"
)

type AdminUserGroupController struct {
}

// AdminUserGroupController 用于用户组相关接口
// 路由注册函数
func RegisterAdminUserGroupRoutes(admin *gin.RouterGroup) {

	ctrl := AdminUserGroupController{}
	// @Summary 获取用户组列表
	// @Description 获取所有用户组信息
	// @Security BearerAuth
	// @Success 200 {object} []models.UserGroup
	// @Router /admin/user_group/list [get]
	admin.GET("/user_group/list", ctrl.ListUserGroup)
	// @Summary 保存用户组
	// @Description 新增或更新用户组信息
	// @Security BearerAuth
	// @Accept json
	// @Param data body models.UserGroup true "用户组信息"
	// @Success 200 {object} map[string]interface{}
	// @Router /admin/user_group/save [post]
	admin.POST("/user_group/save", ctrl.SaveUserGroup)
	// @Summary 删除用户组
	// @Description 根据ID批量删除用户组
	// @Security BearerAuth
	// @Param ids path string true "用户组ID，多个用逗号分隔"
	// @Success 200 {object} string
	// @Router /admin/user_group/delete/{ids} [post]
	admin.POST("/user_group/delete/:ids", ctrl.DeleteUserGroup)
	// @Summary 用户组选项列表
	// @Description 获取所有用户组的选项（仅ID和名称）
	// @Security BearerAuth
	// @Success 200 {object} []map[string]string
	// @Router /admin/user_group/option_list [get]
	admin.GET("/user_group/option_list", ctrl.GroupOptionList)

}
func (a *AdminUserGroupController) ListUserGroup(c *gin.Context) {
	params := dao.BuildParams(c)
	m := &models.UserGroup{}

	items, total, err := m.List(params)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonListWithTotal(c, total, items)
}
func (a *AdminUserGroupController) SaveUserGroup(c *gin.Context) {

	params := dao.BuildParams(c)
	m := models.UserGroup{}
	err := c.ShouldBindJSON(&m)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	err = m.Save(params)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	// 清除用户组的缓存
	service.UserService().ClearCacheByKey(m.GroupName)
	amis.WriteJsonData(c, gin.H{
		"id": m.ID,
	})
}
func (a *AdminUserGroupController) DeleteUserGroup(c *gin.Context) {
	ids := c.Param("ids")

	params := dao.BuildParams(c)
	m := &models.UserGroup{}

	err := m.Delete(params, ids)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	// 清除用户组的缓存
	service.UserService().ClearCacheByKey(m.GroupName)
	amis.WriteJsonOK(c)
}

func (a *AdminUserGroupController) GroupOptionList(c *gin.Context) {
	params := dao.BuildParams(c)
	m := &models.UserGroup{}
	items, _, err := m.List(params, func(db *gorm.DB) *gorm.DB {
		return db.Distinct("id,group_name")
	})
	if err != nil {
		amis.WriteJsonData(c, gin.H{
			"options": make([]map[string]string, 0),
		})
		return
	}
	var names []map[string]string
	for _, n := range items {
		names = append(names, map[string]string{
			"label": n.GroupName,
			"value": n.GroupName,
		})
	}
	slice.SortBy(names, func(a, b map[string]string) bool {
		return a["label"] < b["label"]
	})
	amis.WriteJsonData(c, gin.H{
		"options": names,
	})
}
