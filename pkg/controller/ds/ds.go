package ds

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/klog/v2"
)

type Controller struct{}

func RegisterRoutes(api *gin.RouterGroup) {
	ctrl := &Controller{}

	api.POST("/daemonset/ns/:ns/name/:name/revision/:revision/rollout/undo", ctrl.Undo)
	api.GET("/daemonset/ns/:ns/name/:name/rollout/history", ctrl.History)
	api.POST("/daemonset/ns/:ns/name/:name/restart", ctrl.Restart)
	api.POST("/daemonset/batch/restart", ctrl.BatchRestart)
	api.POST("/daemonset/batch/stop", ctrl.BatchStop)
	api.POST("/daemonset/batch/restore", ctrl.BatchRestore)
}

// @Summary 获取DaemonSet回滚历史
// @Security BearerAuth
// @Param ns path string true "命名空间"
// @Param name path string true "DaemonSet名称"
// @Success 200 {object} string
// @Router /daemonset/ns/{ns}/name/{name}/rollout/history [get]
func (cc *Controller) History(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	list, err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.DaemonSet{}).Namespace(ns).Name(name).
		Ctl().Rollout().History()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, list)
}

// @Summary 重启DaemonSet
// @Security BearerAuth
// @Param ns path string true "命名空间"
// @Param name path string true "DaemonSet名称"
// @Success 200 {object} string
// @Router /daemonset/ns/{ns}/name/{name}/restart [post]
func (cc *Controller) Restart(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	err = kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.DaemonSet{}).Namespace(ns).Name(name).
		Ctl().Rollout().Restart()
	amis.WriteJsonErrorOrOK(c, err)
}

// @Summary 批量重启DaemonSet
// @Security BearerAuth
// @Param body body object true "包含name_list和ns_list的请求体"
// @Success 200 {object} string
// @Router /daemonset/batch/restart [post]
func (cc *Controller) BatchRestart(c *gin.Context) {
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

		x := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.DaemonSet{}).Namespace(ns).Name(name).
			Ctl().Rollout().Restart()
		if x != nil {
			klog.V(6).Infof("批量重启 ds 错误 %s/%s %v", ns, name, x)
			err = x
		}
	}

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

// @Summary 回滚DaemonSet到指定版本
// @Security BearerAuth
// @Param ns path string true "命名空间"
// @Param name path string true "DaemonSet名称"
// @Param revision path string true "回滚版本"
// @Success 200 {object} string
// @Router /daemonset/ns/{ns}/name/{name}/revision/{revision}/rollout/undo [post]
func (cc *Controller) Undo(c *gin.Context) {
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

	result, err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.DaemonSet{}).Namespace(ns).Name(name).
		Ctl().Rollout().Undo(r)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOKMsg(c, result)
}

// @Summary 批量停止DaemonSet
// @Security BearerAuth
// @Param body body object true "包含name_list和ns_list的请求体"
// @Success 200 {object} string
// @Router /daemonset/batch/stop [post]
func (cc *Controller) BatchStop(c *gin.Context) {
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

		x := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.DaemonSet{}).Namespace(ns).Name(name).
			Ctl().DaemonSet().Stop()
		if x != nil {
			klog.V(6).Infof("批量停止 ds 错误 %s/%s %v", ns, name, x)
			err = x
		}
	}

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

// @Summary 批量恢复DaemonSet
// @Security BearerAuth
// @Param body body object true "包含name_list和ns_list的请求体"
// @Success 200 {object} string
// @Router /daemonset/batch/restore [post]
func (cc *Controller) BatchRestore(c *gin.Context) {
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

		x := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.DaemonSet{}).Namespace(ns).Name(name).
			Ctl().DaemonSet().Restore()
		if x != nil {
			klog.V(6).Infof("批量恢复 ds 错误 %s/%s %v", ns, name, x)
			err = x
		}
	}

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}
