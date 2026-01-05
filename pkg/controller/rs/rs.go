package rs

import (
	"fmt"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/response"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
)

type Controller struct{}

// RegisterRoutes 注册 ReplicaSet 相关路由

func RegisterRoutes(r chi.Router) {
	ctrl := &Controller{}
	r.Post("/replicaset/ns/{ns}/name/{name}/restart", response.Adapter(ctrl.Restart))
	r.Post("/replicaset/batch/restart", response.Adapter(ctrl.BatchRestart))
	r.Post("/replicaset/batch/stop", response.Adapter(ctrl.BatchStop))
	r.Post("/replicaset/batch/restore", response.Adapter(ctrl.BatchRestore))
	r.Get("/replicaset/ns/{ns}/name/{name}/events/all", response.Adapter(ctrl.Event))
	r.Get("/replicaset/ns/{ns}/name/{name}/hpa", response.Adapter(ctrl.HPA))
}

// Restart 重启指定的ReplicaSet
// @Summary 重启指定的ReplicaSet
// @Security BearerAuth
// @Param cluster path string true "集群名称"
// @Param ns path string true "命名空间"
// @Param name path string true "ReplicaSet名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/replicaset/ns/{ns}/name/{name}/restart [post]
func (cc *Controller) Restart(c *response.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	err = kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.ReplicaSet{}).Namespace(ns).Name(name).
		Ctl().Rollout().Restart()
	amis.WriteJsonErrorOrOK(c, err)
}

// BatchRestart 批量重启ReplicaSet
// @Summary 批量重启ReplicaSet
// @Security BearerAuth
// @Param cluster path string true "集群名称"
// @Param body body object true "包含name_list和ns_list的请求体"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/replicaset/batch/restart [post]
func (cc *Controller) BatchRestart(c *response.Context) {
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

		x := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.ReplicaSet{}).Namespace(ns).Name(name).
			Ctl().Rollout().Restart()
		if x != nil {
			klog.V(6).Infof("批量重启 rs 错误 %s/%s %v", ns, name, x)
			err = x
		}
	}

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

// BatchStop 批量停止ReplicaSet
// @Summary 批量停止ReplicaSet
// @Security BearerAuth
// @Param cluster path string true "集群名称"
// @Param body body object true "包含name_list和ns_list的请求体"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/replicaset/batch/stop [post]
func (cc *Controller) BatchStop(c *response.Context) {
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

		x := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.ReplicaSet{}).Namespace(ns).Name(name).
			Ctl().Scaler().Stop()
		if x != nil {
			klog.V(6).Infof("批量停止 rs 错误 %s/%s %v", ns, name, x)
			err = x
		}
	}

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

// BatchRestore 批量恢复ReplicaSet
// @Summary 批量恢复ReplicaSet
// @Security BearerAuth
// @Param cluster path string true "集群名称"
// @Param body body object true "包含name_list和ns_list的请求体"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/replicaset/batch/restore [post]
func (cc *Controller) BatchRestore(c *response.Context) {
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

		x := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.ReplicaSet{}).Namespace(ns).Name(name).
			Ctl().Scaler().Restore()
		if x != nil {
			klog.V(6).Infof("批量恢复 rs 错误 %s/%s %v", ns, name, x)
			err = x
		}
	}

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

// Event 获取ReplicaSet相关事件列表
// @Summary 获取ReplicaSet相关事件列表
// @Security BearerAuth
// @Param cluster path string true "集群名称"
// @Param ns path string true "命名空间"
// @Param name path string true "ReplicaSet名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/replicaset/ns/{ns}/name/{name}/events/all [get]
func (cc *Controller) Event(c *response.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var metas []string

	metas = append(metas, name)
	var rs *v1.ReplicaSet
	// 先取rs
	err = kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.ReplicaSet{}).
		Namespace(ns).Name(name).Get(&rs).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	metas = append(metas, rs.ObjectMeta.Name)
	// 再取Pod
	pods, err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Deployment{}).
		Namespace(ns).Name(name).
		Ctl().ReplicaSet().ManagedPods()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	for _, pod := range pods {
		metas = append(metas, pod.ObjectMeta.Name)
	}

	klog.V(6).Infof("meta names = %s", metas)

	var eventList []*unstructured.Unstructured

	sql := kom.Cluster(selectedCluster).
		WithContext(ctx).
		RemoveManagedFields().
		Namespace(ns).
		GVK("events.k8s.io", "v1", "Event")
	// 拼接sql 条件

	// regarding.name = 'x' or regarding.name = 'y'
	var conditions []string
	for _, meta := range metas {
		conditions = append(conditions, fmt.Sprintf("regarding.name = '%s'", meta))
	}
	condStr := strings.Join(conditions, " or ")
	if len(metas) > 0 {
		sql = sql.Where(condStr)
	}

	err = sql.List(&eventList).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, eventList)
}

// HPA 获取ReplicaSet相关HPA列表
// @Summary 获取ReplicaSet相关HPA列表
// @Security BearerAuth
// @Param cluster path string true "集群名称"
// @Param ns path string true "命名空间"
// @Param name path string true "ReplicaSet名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/replicaset/ns/{ns}/name/{name}/hpa [get]
func (cc *Controller) HPA(c *response.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	hpa, err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.ReplicaSet{}).Namespace(ns).Name(name).
		Ctl().ReplicaSet().HPAList()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, hpa)
}
