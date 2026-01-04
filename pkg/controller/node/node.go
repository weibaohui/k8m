package node

import (
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/response"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

type ActionController struct{}

// 从 gin 切换到 chi，使用 chi.Router 替代 gin.RouterGroup
func RegisterActionRoutes(r chi.Router) {
	ctrl := &ActionController{}
	r.Post("/node/drain/name/{name}", response.Adapter(ctrl.Drain))
	r.Post("/node/cordon/name/{name}", response.Adapter(ctrl.Cordon))
	r.Post("/node/uncordon/name/{name}", response.Adapter(ctrl.UnCordon))
	r.Post("/node/batch/drain", response.Adapter(ctrl.BatchDrain))
	r.Post("/node/batch/cordon", response.Adapter(ctrl.BatchCordon))
	r.Post("/node/batch/uncordon", response.Adapter(ctrl.BatchUnCordon))
}

// @Summary 驱逐指定节点
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param name path string true "节点名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/node/drain/name/{name} [post]
func (nc *ActionController) Drain(c *response.Context) {
	name := c.Param("name")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	err = kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Node{}).Name(name).
		Ctl().Node().Drain()
	amis.WriteJsonErrorOrOK(c, err)
}

// @Summary 隔离指定节点
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param name path string true "节点名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/node/cordon/name/{name} [post]
func (nc *ActionController) Cordon(c *response.Context) {
	name := c.Param("name")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	err = kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Node{}).Name(name).
		Ctl().Node().Cordon()
	amis.WriteJsonErrorOrOK(c, err)
}

// @Summary 解除指定节点隔离
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param name path string true "节点名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/node/uncordon/name/{name} [post]
func (nc *ActionController) UnCordon(c *response.Context) {
	name := c.Param("name")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	err = kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Node{}).Name(name).
		Ctl().Node().UnCordon()
	amis.WriteJsonErrorOrOK(c, err)
}

// @Summary 批量驱逐指定的 Kubernetes 节点
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param name_list body []string true "节点名称列表"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/node/batch/drain [post]
func (nc *ActionController) BatchDrain(c *response.Context) {
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var req struct {
		Names []string `json:"name_list"`
	}
	if err = c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

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

// @Summary 批量将指定的 Kubernetes 节点设置为不可调度（cordon）
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param name_list body []string true "节点名称列表"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/node/batch/cordon [post]
func (nc *ActionController) BatchCordon(c *response.Context) {
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var req struct {
		Names []string `json:"name_list"`
	}
	if err = c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

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

// @Summary 批量解除指定节点的隔离状态（Uncordon），使其重新可调度
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param name_list body []string true "节点名称列表"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/node/batch/uncordon [post]
func (nc *ActionController) BatchUnCordon(c *response.Context) {
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var req struct {
		Names []string `json:"name_list"`
	}
	if err = c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

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
