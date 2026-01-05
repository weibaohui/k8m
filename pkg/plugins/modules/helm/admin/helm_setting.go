package admin

import (
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/plugins/modules/helm/models"
	"github.com/weibaohui/k8m/pkg/response"
)

type SettingController struct{}

func (s *SettingController) GetSetting(c *response.Context) {
	setting, err := models.GetOrCreateHelmSetting()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, setting)
}

func (s *SettingController) UpdateSetting(c *response.Context) {
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
