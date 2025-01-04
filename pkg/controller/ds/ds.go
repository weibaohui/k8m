package ds

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/apps/v1"
)

func History(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	ctx := c.Request.Context()
	selectedCluster := amis.GetselectedCluster(c)

	list, err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.DaemonSet{}).Namespace(ns).Name(name).
		Ctl().Rollout().History()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, list)
}
func Restart(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	ctx := c.Request.Context()
	selectedCluster := amis.GetselectedCluster(c)

	err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.DaemonSet{}).Namespace(ns).Name(name).
		Ctl().Rollout().Restart()
	amis.WriteJsonErrorOrOK(c, err)
}
func Undo(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	revision := c.Param("revision")
	ctx := c.Request.Context()
	r := utils.ToInt(revision)
	selectedCluster := amis.GetselectedCluster(c)

	result, err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.DaemonSet{}).Namespace(ns).Name(name).
		Ctl().Rollout().Undo(r)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOKMsg(c, result)
}
