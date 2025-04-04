package config

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/flag"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/service"
)

func Config(c *gin.Context) {
	key := c.Param("key")
	cfg := flag.Init()
	s := ""
	switch key {
	case "AnySelect":
		s = fmt.Sprintf("%v", cfg.AnySelect)
	}
	amis.WriteJsonData(c, s)
}

func GetConfig(c *gin.Context) {
	config, err := service.ConfigService().GetConfig(c.Request.Context())
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, config)
}

func UpdateConfig(c *gin.Context) {
	var config models.Config
	if err := c.ShouldBindJSON(&config); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	if err := service.ConfigService().UpdateConfig(c.Request.Context(), &config); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}
