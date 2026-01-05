package config

import (
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/response"
	"github.com/weibaohui/k8m/pkg/service"
)

type Controller struct {
}

func RegisterConfigRoutes(r chi.Router) {
	ctrl := &Controller{}
	r.Get("/config/all", response.Adapter(ctrl.All))
	r.Post("/config/update", response.Adapter(ctrl.Update))
}

// @Summary 获取系统配置
// @Security BearerAuth
// @Success 200 {object} string
// @Router /admin/config/all [get]
func (cc *Controller) All(c *response.Context) {
	config, err := service.ConfigService().GetConfig()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, config)
}

// @Summary 更新系统配置
// @Security BearerAuth
// @Param config body models.Config true "配置信息"
// @Success 200 {object} string
// @Router /admin/config/update [post]
func (cc *Controller) Update(c *response.Context) {
	var config models.Config
	if err := c.ShouldBindJSON(&config); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	if config.EnableAI == false {
		config.AnySelect = false
	}

	if err := service.ConfigService().UpdateConfig(&config); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}
