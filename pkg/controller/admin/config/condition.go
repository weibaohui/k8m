package config

import (
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/response"
)

type ConditionController struct {
}

// RegisterConditionRoutes 注册路由
// 从 gin 切换到 chi，使用 chi.Router 替代 gin.RouterGroup
func RegisterConditionRoutes(r chi.Router) {
	ctrl := &ConditionController{}
	r.Get("/condition/list", response.Adapter(ctrl.List))
	r.Post("/condition/save", response.Adapter(ctrl.Save))
	r.Post("/condition/delete/{ids}", response.Adapter(ctrl.Delete))
	r.Post("/condition/save/id/{id}/status/{status}", response.Adapter(ctrl.QuickSave))
}

// @Summary 获取条件列表
// @Security BearerAuth
// @Success 200 {object} string
// @Router /admin/condition/list [get]
func (cc *ConditionController) List(c *response.Context) {
	params := dao.BuildParams(c)
	m := &models.ConditionReverse{}

	items, total, err := m.List(params)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonListWithTotal(c, total, items)
}

// @Summary 创建或更新条件
// @Security BearerAuth
// @Success 200 {object} string
// @Router /admin/condition/save [post]
func (cc *ConditionController) Save(c *response.Context) {
	params := dao.BuildParams(c)
	m := models.ConditionReverse{}
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
	amis.WriteJsonData(c, response.H{
		"id": m.ID,
	})
}

// @Summary 删除条件
// @Security BearerAuth
// @Param ids path string true "条件ID，多个用逗号分隔"
// @Success 200 {object} string
// @Router /admin/condition/delete/{ids} [post]
func (cc *ConditionController) Delete(c *response.Context) {
	ids := c.Param("ids")
	params := dao.BuildParams(c)
	m := &models.ConditionReverse{}

	err := m.Delete(params, ids)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

// @Summary 快速保存条件状态
// @Security BearerAuth
// @Param id path int true "条件ID"
// @Param status path string true "状态，例如：true、false"
// @Success 200 {object} string
// @Router /admin/condition/save/id/{id}/status/{status} [post]
func (cc *ConditionController) QuickSave(c *response.Context) {
	id := c.Param("id")
	status := c.Param("status")

	var entity models.ConditionReverse
	entity.ID = utils.ToUInt(id)

	if status == "true" {
		entity.Enabled = true
	} else {
		entity.Enabled = false
	}
	err := dao.DB().Model(&entity).Select("enabled").Updates(entity).Error

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonErrorOrOK(c, err)
}
