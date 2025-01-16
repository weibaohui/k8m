package node

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

func Drain(c *gin.Context) {
	name := c.Param("name")
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Node{}).Name(name).
		Ctl().Node().Drain()
	amis.WriteJsonErrorOrOK(c, err)
}
func Cordon(c *gin.Context) {
	name := c.Param("name")
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Node{}).Name(name).
		Ctl().Node().Cordon()
	amis.WriteJsonErrorOrOK(c, err)
}
func Usage(c *gin.Context) {
	name := c.Param("name")
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	usage := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Node{}).Name(name).
		Ctl().Node().ResourceUsageTable()
	amis.WriteJsonData(c, usage)
}
func UnCordon(c *gin.Context) {
	name := c.Param("name")
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Node{}).Name(name).
		Ctl().Node().UnCordon()
	amis.WriteJsonErrorOrOK(c, err)
}

func BatchDrain(c *gin.Context) {
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	var req struct {
		Names []string `json:"name_list"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var err error
	for i := 0; i < len(req.Names); i++ {
		name := req.Names[i]
		x := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Node{}).Name(name).
			Ctl().Node().Drain()
		if x != nil {
			klog.V(6).Infof("批量驱逐节点错误 %s %v", name, x)
			err = x
		}
	}

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

func BatchCordon(c *gin.Context) {
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	var req struct {
		Names []string `json:"name_list"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var err error
	for i := 0; i < len(req.Names); i++ {
		name := req.Names[i]
		x := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Node{}).Name(name).
			Ctl().Node().Cordon()
		if x != nil {
			klog.V(6).Infof("批量隔离节点错误 %s %v", name, x)
			err = x
		}
	}

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

func BatchUnCordon(c *gin.Context) {
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	var req struct {
		Names []string `json:"name_list"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var err error
	for i := 0; i < len(req.Names); i++ {
		name := req.Names[i]
		x := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Node{}).Name(name).
			Ctl().Node().UnCordon()
		if x != nil {
			klog.V(6).Infof("批量解除节点隔离错误 %s %v", name, x)
			err = x
		}
	}

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}
