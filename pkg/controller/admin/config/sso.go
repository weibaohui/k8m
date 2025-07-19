package config

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	"gorm.io/gorm"
)

type SSOConfigController struct {
}

func RegisterSSOConfigRoutes(admin *gin.RouterGroup) {
	ctrl := &SSOConfigController{}
	// SSO 配置
	admin.GET("/config/sso/list", ctrl.List)
	admin.POST("/config/sso/save", ctrl.Save)
	admin.POST("/config/sso/delete/:ids", ctrl.Delete)
	admin.POST("/config/sso/save/id/:id/status/:enabled", ctrl.QuickSave)
}

func (sc *SSOConfigController) List(c *gin.Context) {
	params := dao.BuildParams(c)
	m := &models.SSOConfig{}

	items, total, err := m.List(params)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonListWithTotal(c, total, items)
}

func (sc *SSOConfigController) Save(c *gin.Context) {
	params := dao.BuildParams(c)
	m := models.SSOConfig{}
	err := c.ShouldBindJSON(&m)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	err = m.Save(params, func(db *gorm.DB) *gorm.DB {
		return db.Select([]string{"name", "type", "client_id", "client_secret", "issuer", "prefer_user_name_keys", "scopes"})
	})
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, gin.H{
		"id": m.ID,
	})
}

func (sc *SSOConfigController) Delete(c *gin.Context) {
	ids := c.Param("ids")
	params := dao.BuildParams(c)
	m := &models.SSOConfig{}

	err := m.Delete(params, ids)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

func (sc *SSOConfigController) QuickSave(c *gin.Context) {
	id := c.Param("id")
	enabled := c.Param("enabled")

	var entity models.SSOConfig
	entity.ID = utils.ToUInt(id)

	if enabled == "true" {
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
