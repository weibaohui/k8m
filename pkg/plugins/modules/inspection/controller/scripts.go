package controller

import (
	"github.com/duke-git/lancet/v2/slice"
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/inspection/models"
	"github.com/weibaohui/k8m/pkg/response"
	"k8s.io/klog/v2"
)

type AdminLuaScriptController struct {
}

func RegisterAdminLuaScriptRoutes(arg *chi.Router) {
	admin := arg.Group("/plugins/" + modules.PluginNameInspection)
	ctrl := &AdminLuaScriptController{}
	admin.GET("/script/list", ctrl.LuaScriptList)
	admin.POST("/script/delete/:ids", ctrl.LuaScriptDelete)
	admin.POST("/script/save", ctrl.LuaScriptSave)
	admin.POST("/script/load", ctrl.LuaScriptLoad)
	admin.GET("/script/option_list", ctrl.LuaScriptOptionList)
}

// @Summary 获取Lua脚本列表
// @Security BearerAuth
// @Success 200 {object} string
// @Router /admin/plugins/inspection/script/list [get]
func (s *AdminLuaScriptController) LuaScriptList(c *response.Context) {
	params := dao.BuildParams(c)
	m := &models.InspectionLuaScript{}

	items, total, err := m.List(params)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonListWithTotal(c, total, items)
}

// @Summary 保存Lua脚本
// @Security BearerAuth
// @Success 200 {object} string
// @Router /admin/plugins/inspection/script/save [post]
func (s *AdminLuaScriptController) LuaScriptSave(c *response.Context) {
	params := dao.BuildParams(c)
	m := models.InspectionLuaScript{}
	err := c.ShouldBindJSON(&m)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	if m.ScriptType == "" {
		m.ScriptType = constants.LuaScriptTypeCustom
	}

	err = m.Save(params)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonOK(c)
}

// @Summary 删除Lua脚本
// @Security BearerAuth
// @Param ids path string true "脚本ID，多个用逗号分隔"
// @Success 200 {object} string
// @Router /admin/plugins/inspection/script/delete/{ids} [post]
func (s *AdminLuaScriptController) LuaScriptDelete(c *response.Context) {
	ids := c.Param("ids")
	params := dao.BuildParams(c)
	params.UserName = ""

	m := &models.InspectionLuaScript{}

	err := m.Delete(params, ids)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

// @Summary 获取Lua脚本选项列表
// @Security BearerAuth
// @Success 200 {object} string
// @Router /admin/plugins/inspection/script/option_list [get]
func (s *AdminLuaScriptController) LuaScriptOptionList(c *response.Context) {
	m := models.InspectionLuaScript{}
	params := dao.BuildParams(c)
	params.PerPage = 100000
	list, _, err := m.List(params)

	if err != nil {
		amis.WriteJsonData(c, response.H{
			"options": make([]map[string]string, 0),
		})
		return
	}
	var scripts []map[string]string
	for _, n := range list {
		scripts = append(scripts, map[string]string{
			"label":       n.Name,
			"value":       n.ScriptCode,
			"script_code": n.ScriptCode,
			"name":        n.Name,
			"description": n.Description,
		})
	}
	slice.SortBy(scripts, func(a, b map[string]string) bool {
		return a["label"] < b["label"]
	})
	amis.WriteJsonData(c, response.H{
		"options": scripts,
	})
}

// @Summary 加载内置Lua脚本
// @Security BearerAuth
// @Success 200 {object} string
// @Router /admin/plugins/inspection/script/load [post]
func (s *AdminLuaScriptController) LuaScriptLoad(c *response.Context) {
	// 删除后，重新插入内置脚本
	err := dao.DB().Model(&models.InspectionLuaScript{}).Where("script_type = ?", constants.LuaScriptTypeBuiltin).Delete(&models.InspectionLuaScript{}).Error
	if err != nil {
		klog.Errorf("删除内置巡检脚本失败: %v", err)
		amis.WriteJsonError(c, err)
		return
	}
	err = dao.DB().Model(&models.InspectionLuaScript{}).CreateInBatches(models.BuiltinLuaScripts, 100).Error
	if err != nil {
		klog.Errorf("插入内置巡检脚本失败: %v", err)
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}
