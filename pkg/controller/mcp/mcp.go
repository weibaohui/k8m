package mcp

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/service"
)

func List(c *gin.Context) {
	servers := service.McpService().Host().ListServers()
	amis.WriteJsonData(c, servers)
}
func Connect(c *gin.Context) {
	name := c.Param("name")
	err := service.McpService().Host().ConnectServer(c.Request.Context(), name)
	amis.WriteJsonErrorOrOK(c, err)
}
