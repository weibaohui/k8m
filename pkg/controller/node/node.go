package node

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/service"
)

func Drain(c *gin.Context) {
	name := c.Param("name")
	ctx := c.Request.Context()
	err := service.NodeService().Drain(ctx, name)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}
func Cordon(c *gin.Context) {
	name := c.Param("name")
	ctx := c.Request.Context()
	err := service.NodeService().Cordon(ctx, name)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}
func UnCordon(c *gin.Context) {
	name := c.Param("name")
	ctx := c.Request.Context()
	err := service.NodeService().UnCordon(ctx, name)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}
