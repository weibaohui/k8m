package cronjob

import (
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/response"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/batch/v1"
	"k8s.io/klog/v2"
)

type Controller struct{}

// 从 gin 切换到 chi，使用 chi.Router 替代 gin.RouterGroup
func RegisterRoutes(r chi.Router) {
	ctrl := &Controller{}
	r.Post("/cronjob/pause/ns/{ns}/name/{name}", response.Adapter(ctrl.Pause))
	r.Post("/cronjob/resume/ns/{ns}/name/{name}", response.Adapter(ctrl.Resume))
	r.Post("/cronjob/batch/resume", response.Adapter(ctrl.BatchResume))
	r.Post("/cronjob/batch/pause", response.Adapter(ctrl.BatchPause))
}

// @Summary 暂停 CronJob
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param ns path string true "命名空间"
// @Param name path string true "CronJob 名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/cronjob/pause/ns/{ns}/name/{name} [post]
func (cc *Controller) Pause(c *response.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	err = kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.CronJob{}).Namespace(ns).Name(name).
		Ctl().CronJob().Pause()
	amis.WriteJsonErrorOrOK(c, err)
}

// @Summary 恢复 CronJob
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param ns path string true "命名空间"
// @Param name path string true "CronJob 名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/cronjob/resume/ns/{ns}/name/{name} [post]
func (cc *Controller) Resume(c *response.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	err = kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.CronJob{}).Namespace(ns).Name(name).
		Ctl().CronJob().Resume()
	amis.WriteJsonErrorOrOK(c, err)
}

// @Summary 批量恢复 CronJob
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param request body object true "批量恢复请求体，包含 name_list 和 ns_list"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/cronjob/batch/resume [post]
func (cc *Controller) BatchResume(c *response.Context) {
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

		x := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.CronJob{}).Namespace(ns).Name(name).
			Ctl().CronJob().Resume()
		if x != nil {
			klog.V(6).Infof("批量恢复 cronjob 错误 %s/%s %v", ns, name, x)
			err = x
		}
	}

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

// @Summary 批量暂停 CronJob
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param request body object true "批量暂停请求体，包含 name_list 和 ns_list"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/cronjob/batch/pause [post]
func (cc *Controller) BatchPause(c *response.Context) {
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
	if err = c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	for i := 0; i < len(req.Names); i++ {
		name := req.Names[i]
		ns := req.Namespaces[i]

		x := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.CronJob{}).Namespace(ns).Name(name).
			Ctl().CronJob().Pause()
		if x != nil {
			klog.V(6).Infof("批量暂停 cronjob 错误 %s/%s %v", ns, name, x)
			err = x
		}
	}

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}
