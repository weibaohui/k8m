package param

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/flag"
)

// Config 获取某一个参数配置
func Config(c *gin.Context) {
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
