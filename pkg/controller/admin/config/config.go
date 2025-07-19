package config

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/service"
)

type Controller struct {
}

func RegisterConfigRoutes(admin *gin.RouterGroup) {
	ctrl := &Controller{}
	admin.GET("/config/all", ctrl.All)
	admin.POST("/config/update", ctrl.Update)
}
func (cc *Controller) All(c *gin.Context) {
	config, err := service.ConfigService().GetConfig()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, config)
}

func (cc *Controller) Update(c *gin.Context) {
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
	_ = service.ConfigService().UpdateFlagFromDBConfig()
	amis.WriteJsonOK(c)
}
