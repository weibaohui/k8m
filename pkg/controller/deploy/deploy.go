package deploy

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/kubectl"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
)

func UpdateImageTag(c *gin.Context) {
	var ns = c.Param("ns")
	var name = c.Param("name")
	var tag = c.Param("tag")
	var containerName = c.Param("container_name")
	ctx := c.Request.Context()
	deploy, _ := kubectl.Init().UpdateDeployImageTag(ctx, ns, name, containerName, tag)
	amis.WriteJsonData(c, deploy)

}
func Restart(c *gin.Context) {
	var ns = c.Param("ns")
	var name = c.Param("name")
	ctx := c.Request.Context()
	deploy, _ := kubectl.Init().RestartDeploy(ctx, ns, name)
	amis.WriteJsonData(c, deploy)
}
