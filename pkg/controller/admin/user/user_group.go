package user

import (
	"github.com/duke-git/lancet/v2/slice"
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/response"
	"github.com/weibaohui/k8m/pkg/service"
	"gorm.io/gorm"
)

type AdminUserGroupController struct {
}

// AdminUserGroupController 用于用户组相关接口

func RegisterAdminUserGroupRoutes(r chi.Router) {

	ctrl := AdminUserGroupController{}
	r.Get("/user_group/list", response.Adapter(ctrl.ListUserGroup))
	r.Post("/user_group/save", response.Adapter(ctrl.SaveUserGroup))
	r.Post("/user_group/save_menu", response.Adapter(ctrl.SaveUserGroupMenu))
	r.Post("/user_group/delete/{ids}", response.Adapter(ctrl.DeleteUserGroup))
	r.Get("/user_group/option_list", response.Adapter(ctrl.GroupOptionList))
}

// @Summary 获取用户组列表
// @Description 获取所有用户组信息
// @Security BearerAuth
// @Success 200 {object} []models.UserGroup
// @Router /admin/user_group/list [get]
func (a *AdminUserGroupController) ListUserGroup(c *response.Context) {
	params := dao.BuildParams(c)
	m := &models.UserGroup{}

	items, total, err := m.List(params)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonListWithTotal(c, total, items)
}

// @Summary 保存用户组
// @Description 新增或更新用户组信息
// @Security BearerAuth
// @Accept json
// @Param data body models.UserGroup true "用户组信息"
// @Success 200 {object} map[string]interface{}
// @Router /admin/user_group/save [post]
func (a *AdminUserGroupController) SaveUserGroup(c *response.Context) {

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
	amis.WriteJsonData(c, response.H{
		"id": m.ID,
	})
}

// @Summary 保存用户组菜单配置
// @Description 保存用户组的菜单配置数据
// @Security BearerAuth
// @Accept json
// @Param data body map[string]interface{} true "菜单配置信息"
// @Success 200 {object} map[string]interface{}
// @Router /admin/user_group/save_menu [post]
func (a *AdminUserGroupController) SaveUserGroupMenu(c *response.Context) {
	var requestData struct {
		ID       uint   `json:"id" binding:"required"`
		MenuData string `json:"menu_data"`
	}

	err := c.ShouldBindJSON(&requestData)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	params := dao.BuildParams(c)
	m := &models.UserGroup{}

	// 先查找用户组
	userGroup, err := m.GetOne(params, func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", requestData.ID)
	})
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 更新菜单数据
	userGroup.MenuData = requestData.MenuData
	err = userGroup.Save(params)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 清除用户组的缓存
	service.UserService().ClearCacheByKey(userGroup.GroupName)
	amis.WriteJsonOK(c)
}

// @Summary 删除用户组
// @Description 根据ID批量删除用户组
// @Security BearerAuth
// @Param ids path string true "用户组ID，多个用逗号分隔"
// @Success 200 {object} string
// @Router /admin/user_group/delete/{ids} [post]
func (a *AdminUserGroupController) DeleteUserGroup(c *response.Context) {
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

// @Summary 用户组选项列表
// @Description 获取所有用户组的选项（仅ID和名称）
// @Security BearerAuth
// @Success 200 {object} []map[string]string
// @Router /admin/user_group/option_list [get]
func (a *AdminUserGroupController) GroupOptionList(c *response.Context) {
	params := dao.BuildParams(c)
	m := &models.UserGroup{}
	items, _, err := m.List(params, func(db *gorm.DB) *gorm.DB {
		return db.Distinct("id,group_name")
	})
	if err != nil {
		amis.WriteJsonData(c, response.H{
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
	amis.WriteJsonData(c, response.H{
		"options": names,
	})
}
