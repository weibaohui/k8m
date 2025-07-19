package config

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
)

type ConditionController struct {
}

func RegisterConditionRoutes(admin *gin.RouterGroup) {
	ctrl := &ConditionController{}
	admin.GET("/condition/list", ctrl.List)
	admin.POST("/condition/save", ctrl.Save)
	admin.POST("/condition/delete/:ids", ctrl.Delete)
	admin.POST("/condition/save/id/:id/status/:status", ctrl.QuickSave)
}

func (cc *ConditionController) List(c *gin.Context) {
	params := dao.BuildParams(c)
	m := &models.ConditionReverse{}

	items, total, err := m.List(params)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonListWithTotal(c, total, items)
}

func (cc *ConditionController) Save(c *gin.Context) {
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
	amis.WriteJsonData(c, gin.H{
		"id": m.ID,
	})
}

func (cc *ConditionController) Delete(c *gin.Context) {
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

func (cc *ConditionController) QuickSave(c *gin.Context) {
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
