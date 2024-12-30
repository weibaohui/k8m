package node

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
)

func Drain(c *gin.Context) {
	name := c.Param("name")
	ctx := c.Request.Context()
	err := kom.DefaultCluster().WithContext(ctx).Resource(&v1.Node{}).Name(name).
		Ctl().Node().Drain()
	amis.WriteJsonErrorOrOK(c, err)
}
func Cordon(c *gin.Context) {
	name := c.Param("name")
	ctx := c.Request.Context()
	err := kom.DefaultCluster().WithContext(ctx).Resource(&v1.Node{}).Name(name).
		Ctl().Node().Cordon()
	amis.WriteJsonErrorOrOK(c, err)
}
func Usage(c *gin.Context) {
	name := c.Param("name")
	ctx := c.Request.Context()
	usage := kom.DefaultCluster().WithContext(ctx).Resource(&v1.Node{}).Name(name).
		Ctl().Node().ResourceUsageTable()
	amis.WriteJsonData(c, usage)
}
func UnCordon(c *gin.Context) {
	name := c.Param("name")
	ctx := c.Request.Context()
	err := kom.DefaultCluster().WithContext(ctx).Resource(&v1.Node{}).Name(name).
		Ctl().Node().UnCordon()
	amis.WriteJsonErrorOrOK(c, err)
}
