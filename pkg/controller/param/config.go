package param

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/flag"
)

// Config 获取某一个参数配置
// @Summary 获取配置项
// @Description 获取指定key的系统配置项
// @Security BearerAuth
// @Param key path string true "配置项key"
// @Success 200 {object} string
// @Router /params/config/{key} [get]
func (pc *Controller) Config(c *gin.Context) {
	key := c.Param("key")
	cfg := flag.Init()
	s := ""
	switch key {
	case "AnySelect":
		s = fmt.Sprintf("%v", cfg.AnySelect)
	case "ProductName":
		s = fmt.Sprintf("%v", cfg.ProductName)
	}
	amis.WriteJsonData(c, s)
}
