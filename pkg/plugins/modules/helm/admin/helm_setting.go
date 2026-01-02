package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/plugins/modules/helm/models"
)

type SettingController struct{}

func (s *SettingController) GetSetting(c *gin.Context) {
	setting, err := models.GetOrCreateHelmSetting()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, setting)
}

func (s *SettingController) UpdateSetting(c *gin.Context) {
	var in models.HelmSetting
	if err := c.ShouldBindJSON(&in); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	if _, err := models.UpdateHelmSetting(&in); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}
