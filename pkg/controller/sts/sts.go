package sts

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/klog/v2"
)

func History(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	list, err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.StatefulSet{}).Namespace(ns).Name(name).
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
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	err = kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.StatefulSet{}).Namespace(ns).Name(name).
		Ctl().Rollout().Restart()
	amis.WriteJsonErrorOrOK(c, err)
}

func BatchRestart(c *gin.Context) {
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var req struct {
		Names      []string `json:"name_list"`
		Namespaces []string `json:"ns_list"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	for i := 0; i < len(req.Names); i++ {
		name := req.Names[i]
		ns := req.Namespaces[i]

		x := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.StatefulSet{}).Namespace(ns).Name(name).
			Ctl().Rollout().Restart()
		if x != nil {
			klog.V(6).Infof("批量重启 sts 错误 %s/%s %v", ns, name, x)
			err = x
		}
	}

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

func BatchStop(c *gin.Context) {
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var req struct {
		Names      []string `json:"name_list"`
		Namespaces []string `json:"ns_list"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	for i := 0; i < len(req.Names); i++ {
		name := req.Names[i]
		ns := req.Namespaces[i]

		x := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.StatefulSet{}).Namespace(ns).Name(name).
			Ctl().Scaler().Stop()
		if x != nil {
			klog.V(6).Infof("批量停止 sts 错误 %s/%s %v", ns, name, x)
			err = x
		}
	}

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

func BatchRestore(c *gin.Context) {
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var req struct {
		Names      []string `json:"name_list"`
		Namespaces []string `json:"ns_list"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	for i := 0; i < len(req.Names); i++ {
		name := req.Names[i]
		ns := req.Namespaces[i]

		x := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.StatefulSet{}).Namespace(ns).Name(name).
			Ctl().Scaler().Restore()
		if x != nil {
			klog.V(6).Infof("批量恢复 sts 错误 %s/%s %v", ns, name, x)
			err = x
		}
	}

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}
func Scale(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	replica := c.Param("replica")
	r := utils.ToInt32(replica)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	ctx := amis.GetContextWithUser(c)
	err = kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.StatefulSet{}).
		Namespace(ns).Name(name).
		Ctl().Scaler().Scale(r)
	amis.WriteJsonErrorOrOK(c, err)
}
func Undo(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	revision := c.Param("revision")
	ctx := amis.GetContextWithUser(c)
	r := utils.ToInt(revision)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	result, err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.StatefulSet{}).Namespace(ns).Name(name).
		Ctl().Rollout().Undo(r)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOKMsg(c, result)
}

func HPA(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	hpa, err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.StatefulSet{}).Namespace(ns).Name(name).
		Ctl().StatefulSet().HPAList()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, hpa)
}
