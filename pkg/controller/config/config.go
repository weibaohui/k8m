package config

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/flag"
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
