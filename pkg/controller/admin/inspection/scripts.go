package inspection

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/models"
	"k8s.io/klog/v2"
)

func LuaScriptList(c *gin.Context) {
	params := dao.BuildParams(c)
	m := &models.InspectionLuaScript{}

	items, total, err := m.List(params)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonListWithTotal(c, total, items)
}
func LuaScriptSave(c *gin.Context) {
	params := dao.BuildParams(c)
	m := &models.InspectionLuaScript{}
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
func LuaScriptDelete(c *gin.Context) {
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
func LuaScriptLoad(c *gin.Context) {
	// 删除后，重新插入内置脚本
	err := dao.DB().Model(&models.InspectionLuaScript{}).Where("script_type == ?", constants.LuaScriptTypeBuiltin).Delete(&models.InspectionLuaScript{}).Error
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
