package cronjob

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/batch/v1"
)

func Pause(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	ctx := c.Request.Context()
	err := kom.DefaultCluster().WithContext(ctx).Resource(&v1.CronJob{}).Namespace(ns).Name(name).
		Ctl().CronJob().Pause()
	amis.WriteJsonErrorOrOK(c, err)
}
func Resume(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	ctx := c.Request.Context()
	err := kom.DefaultCluster().WithContext(ctx).Resource(&v1.CronJob{}).Namespace(ns).Name(name).
		Ctl().CronJob().Resume()
	amis.WriteJsonErrorOrOK(c, err)
}
