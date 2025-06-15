package config

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	"gorm.io/gorm"
)

func SSOConfigList(c *gin.Context) {
	params := dao.BuildParams(c)
	m := &models.SSOConfig{}

	items, total, err := m.List(params)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonListWithTotal(c, total, items)
}

func SSOConfigSave(c *gin.Context) {
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

func SSOConfigDelete(c *gin.Context) {
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

func SSOConfigQuickSave(c *gin.Context) {
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
