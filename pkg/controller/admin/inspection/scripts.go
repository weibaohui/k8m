package inspection

import (
	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/models"
	"k8s.io/klog/v2"
)

type AdminLuaScriptController struct {
}

func RegisterAdminLuaScriptRoutes(admin *gin.RouterGroup) {
	ctrl := &AdminLuaScriptController{}
	admin.GET("/inspection/script/list", ctrl.LuaScriptList)
	admin.POST("/inspection/script/delete/:ids", ctrl.LuaScriptDelete)
	admin.POST("/inspection/script/save", ctrl.LuaScriptSave)
	admin.POST("/inspection/script/load", ctrl.LuaScriptLoad)
	admin.GET("/inspection/script/option_list", ctrl.LuaScriptOptionList)

}

func (s *AdminLuaScriptController) LuaScriptList(c *gin.Context) {
	params := dao.BuildParams(c)
	m := &models.InspectionLuaScript{}

	items, total, err := m.List(params)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonListWithTotal(c, total, items)
}
func (s *AdminLuaScriptController) LuaScriptSave(c *gin.Context) {
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
func (s *AdminLuaScriptController) LuaScriptDelete(c *gin.Context) {
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

func (s *AdminLuaScriptController) LuaScriptOptionList(c *gin.Context) {
	m := models.InspectionLuaScript{}
	params := dao.BuildParams(c)
	params.PerPage = 100000
	list, _, err := m.List(params)

	if err != nil {
		amis.WriteJsonData(c, gin.H{
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
	amis.WriteJsonData(c, gin.H{
		"options": scripts,
	})

}

func (s *AdminLuaScriptController) LuaScriptLoad(c *gin.Context) {
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
