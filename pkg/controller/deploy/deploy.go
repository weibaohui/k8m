package deploy

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/apps/v1"
	eventsv1 "k8s.io/api/events/v1"
	"k8s.io/klog/v2"
)

func BatchStop(c *gin.Context) {
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	var req struct {
		Names      []string `json:"name_list"`
		Namespaces []string `json:"ns_list"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var err error
	for i := 0; i < len(req.Names); i++ {
		name := req.Names[i]
		ns := req.Namespaces[i]
		x := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Deployment{}).Namespace(ns).Name(name).
			Ctl().Scaler().Stop()
		if x != nil {
			klog.V(6).Infof("批量停止 deploy 错误 %s/%s %v", ns, name, x)

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
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	var req struct {
		Names      []string `json:"name_list"`
		Namespaces []string `json:"ns_list"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var err error
	for i := 0; i < len(req.Names); i++ {
		name := req.Names[i]
		ns := req.Namespaces[i]
		x := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Deployment{}).Namespace(ns).Name(name).
			Ctl().Scaler().Restore()
		if x != nil {
			klog.V(6).Infof("批量恢复 deploy 错误 %s/%s %v", ns, name, x)

			err = x
		}
	}

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}
func Restart(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Deployment{}).Namespace(ns).Name(name).
		Ctl().Rollout().Restart()
	amis.WriteJsonErrorOrOK(c, err)
}
func BatchRestart(c *gin.Context) {
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	var req struct {
		Names      []string `json:"name_list"`
		Namespaces []string `json:"ns_list"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var err error
	for i := 0; i < len(req.Names); i++ {
		name := req.Names[i]
		ns := req.Namespaces[i]
		x := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Deployment{}).Namespace(ns).Name(name).
			Ctl().Rollout().Restart()
		if x != nil {
			klog.V(6).Infof("批量重启 deploy 错误 %s/%s %v", ns, name, x)

			err = x
		}
	}

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}
func History(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	list, err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Deployment{}).Namespace(ns).Name(name).
		Ctl().Rollout().History()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, list)
}
func Pause(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Deployment{}).Namespace(ns).Name(name).
		Ctl().Rollout().Pause()
	amis.WriteJsonErrorOrOK(c, err)
}
func Resume(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Deployment{}).Namespace(ns).Name(name).
		Ctl().Rollout().Resume()
	amis.WriteJsonErrorOrOK(c, err)
}
func Scale(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	replica := c.Param("replica")
	r := utils.ToInt32(replica)

	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Deployment{}).Namespace(ns).Name(name).
		Ctl().Scaler().Scale(r)
	amis.WriteJsonErrorOrOK(c, err)
}
func Undo(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	revision := c.Param("revision")
	ctx := c.Request.Context()
	r := utils.ToInt(revision)
	selectedCluster := amis.GetSelectedCluster(c)

	result, err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Deployment{}).Namespace(ns).Name(name).
		Ctl().Rollout().Undo(r)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOKMsg(c, result)
}

// Event 显示deploy下所有的事件列表，包括deploy、rs、pod
func Event(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	var metas []string

	metas = append(metas, name)
	// 先取rs
	rs, err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Deployment{}).
		Namespace(ns).Name(name).
		Ctl().Deployment().ManagedLatestReplicaSet()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	metas = append(metas, rs.ObjectMeta.Name)
	// 再取Pod
	pods, err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Deployment{}).
		Namespace(ns).Name(name).
		Ctl().Deployment().ManagedPods()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	for _, pod := range pods {
		metas = append(metas, pod.ObjectMeta.Name)
	}

	klog.V(6).Infof("meta names = %s", metas)

	// 通过mates 获取事件
	var eventList []*eventsv1.Event
	sql := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&eventsv1.Event{})
	sql = sql.AllNamespace()
	for _, meta := range metas {
		sql = sql.Where("regarding.name = ?", meta)
	}
	err = sql.List(&eventList).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, eventList)
}
