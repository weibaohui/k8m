package sts

import (
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/response"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/klog/v2"
)

type Controller struct{}

// RegisterRoutes 注册 StatefulSet 相关路由
// 从 gin 切换到 chi，使用 chi.Router 替代 gin.RouterGroup
func RegisterRoutes(r chi.Router) {
	ctrl := &Controller{}

	r.Post("/statefulset/ns/{ns}/name/{name}/revision/{revision}/rollout/undo", response.Adapter(ctrl.Undo))
	r.Get("/statefulset/ns/{ns}/name/{name}/rollout/history", response.Adapter(ctrl.History))
	r.Post("/statefulset/ns/{ns}/name/{name}/restart", response.Adapter(ctrl.Restart))
	r.Post("/statefulset/batch/restart", response.Adapter(ctrl.BatchRestart))
	r.Post("/statefulset/batch/stop", response.Adapter(ctrl.BatchStop))
	r.Post("/statefulset/batch/restore", response.Adapter(ctrl.BatchRestore))
	r.Post("/statefulset/ns/{ns}/name/{name}/scale/replica/{replica}", response.Adapter(ctrl.Scale))
	r.Get("/statefulset/ns/{ns}/name/{name}/hpa", response.Adapter(ctrl.HPA))

}

// @Summary 获取StatefulSet滚动历史
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param ns path string true "命名空间"
// @Param name path string true "StatefulSet名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/statefulset/ns/{ns}/name/{name}/rollout/history [get]
func (cc *Controller) History(c *response.Context) {
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

// @Summary 重启StatefulSet
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param ns path string true "命名空间"
// @Param name path string true "StatefulSet名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/statefulset/ns/{ns}/name/{name}/restart [post]
func (cc *Controller) Restart(c *response.Context) {
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

// @Summary 批量重启StatefulSet
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param name_list body []string true "StatefulSet名称列表"
// @Param ns_list body []string true "命名空间列表"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/statefulset/batch/restart [post]
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

// @Summary 批量停止StatefulSet
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param name_list body []string true "StatefulSet名称列表"
// @Param ns_list body []string true "命名空间列表"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/statefulset/batch/stop [post]
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

// @Summary 批量恢复StatefulSet
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param name_list body []string true "StatefulSet名称列表"
// @Param ns_list body []string true "命名空间列表"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/statefulset/batch/restore [post]
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

// @Summary 扩缩容StatefulSet
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param ns path string true "命名空间"
// @Param name path string true "StatefulSet名称"
// @Param replica path int true "副本数"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/statefulset/ns/{ns}/name/{name}/scale/replica/{replica} [post]
func (cc *Controller) Scale(c *response.Context) {
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

// @Summary 回滚StatefulSet到指定版本
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param ns path string true "命名空间"
// @Param name path string true "StatefulSet名称"
// @Param revision path int true "版本号"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/statefulset/ns/{ns}/name/{name}/revision/{revision}/rollout/undo [post]
func (cc *Controller) Undo(c *response.Context) {
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

// @Summary 获取StatefulSet的HPA列表
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param ns path string true "命名空间"
// @Param name path string true "StatefulSet名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/statefulset/ns/{ns}/name/{name}/hpa [get]
func (cc *Controller) HPA(c *response.Context) {
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
