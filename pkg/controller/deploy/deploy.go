package deploy

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/apps/v1"
)

func UpdateImageTag(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	var tag = c.Param("tag")
	var containerName = c.Param("container_name")
	ctx := c.Request.Context()

	_, err := kom.DefaultCluster().WithContext(ctx).Resource(&v1.Deployment{}).Namespace(ns).Name(name).
		Ctl().Deployment().ReplaceImageTag(containerName, tag)
	amis.WriteJsonErrorOrOK(c, err)
}
func Restart(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	ctx := c.Request.Context()

	err := kom.DefaultCluster().WithContext(ctx).
		Resource(&v1.Deployment{}).
		Namespace(ns).Name(name).
		Ctl().Rollout().Restart()
	amis.WriteJsonErrorOrOK(c, err)
}
func History(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	ctx := c.Request.Context()
	list, _ := kom.DefaultCluster().WithContext(ctx).
		Resource(&v1.Deployment{}).
		Namespace(ns).Name(name).
		Ctl().Rollout().History()
	amis.WriteJsonData(c, list)
}
