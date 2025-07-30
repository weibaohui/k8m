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

// @Summary 获取系统配置
// @Security BearerAuth
// @Success 200 {object} string
// @Router /admin/config/all [get]
func (cc *Controller) All(c *gin.Context) {
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
