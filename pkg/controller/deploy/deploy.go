package deploy

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/kubectl"
	"github.com/weibaohui/k8m/internal/utils/amis"
)

func UpdateImageTag(c *gin.Context) {
	var ns = c.Param("ns")
	var name = c.Param("name")
	var tag = c.Param("tag")
	var containerName = c.Param("container_name")
	deploy, _ := kubectl.Init().UpdateDeployImageTag(ns, name, containerName, tag)
	amis.WriteJsonData(c, deploy)

}
func Restart(c *gin.Context) {
	var ns = c.Param("ns")
	var name = c.Param("name")
	deploy, _ := kubectl.Init().RestartDeploy(ns, name)
	amis.WriteJsonData(c, deploy)
}
