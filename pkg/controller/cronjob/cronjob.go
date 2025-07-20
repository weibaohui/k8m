package cronjob

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/batch/v1"
	"k8s.io/klog/v2"
)

type Controller struct{}

func RegisterRoutes(api *gin.RouterGroup) {
	ctrl := &Controller{}
	api.POST("/cronjob/pause/ns/:ns/name/:name", ctrl.Pause)
	api.POST("/cronjob/resume/ns/:ns/name/:name", ctrl.Resume)
	api.POST("/cronjob/batch/resume", ctrl.BatchResume)
	api.POST("/cronjob/batch/pause", ctrl.BatchPause)
}

// @Summary 暂停 CronJob
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param ns path string true "命名空间"
// @Param name path string true "CronJob 名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/cronjob/pause/ns/{ns}/name/{name} [post]
func (cc *Controller) Pause(c *gin.Context) {
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
func (cc *Controller) Resume(c *gin.Context) {
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
func (cc *Controller) BatchResume(c *gin.Context) {
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
func (cc *Controller) BatchPause(c *gin.Context) {
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
