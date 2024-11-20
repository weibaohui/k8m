package deploy

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/service"
)

func UpdateImageTag(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	var tag = c.Param("tag")
	var containerName = c.Param("container_name")
	ctx := c.Request.Context()
	deployService := service.DeploymentService()
	deploy, _ := deployService.UpdateDeployImageTag(ctx, ns, name, containerName, tag)
	amis.WriteJsonData(c, deploy)

}
func Restart(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	ctx := c.Request.Context()
	deployService := service.DeploymentService()
	deploy, _ := deployService.RestartDeploy(ctx, ns, name)
	amis.WriteJsonData(c, deploy)
}
