package dynamic

import (
	"context"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/response"
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
)

type PodLinkController struct{}

func RegisterPodLinkRoutes(api *chi.Router) {
	ctrl := &PodLinkController{}
	api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name/links/services", ctrl.LinksServices)
	api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name/links/endpoints", ctrl.LinksEndpoints)
	api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name/links/pvc", ctrl.LinksPVC)
	api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name/links/pv", ctrl.LinksPV)
	api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name/links/ingress", ctrl.LinksIngress)
	api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name/links/env", ctrl.LinksEnv)
	api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name/links/envFromPod", ctrl.LinksEnvFromPod)
	api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name/links/configmap", ctrl.LinksConfigMap)
	api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name/links/secret", ctrl.LinksSecret)
	api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name/links/node", ctrl.LinksNode)
	api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name/links/pod", ctrl.LinksPod)

}

var linkCacheTTL = 3 * time.Second

func getPod(selectedCluster string, ctx context.Context, ns string, name string, kind string, group string, version string) (*v1.Pod, error) {
	var pod *v1.Pod
	var err error
	kk := kom.Cluster(selectedCluster).WithContext(ctx).
		CRD(group, version, kind).
		Namespace(ns).
		Name(name).
		WithCache(linkCacheTTL)
	pod, err = kk.Ctl().CRD().ManagedPod()

	if err == nil && pod != nil {
		return pod, nil
	}
	switch kind {
	case "Pod":
		err = kk.Get(&pod).Error
	case "Deployment":
		pod, err = kk.Ctl().Deployment().ManagedPod()
	case "StatefulSet":
		pod, err = kk.Ctl().StatefulSet().ManagedPod()
	case "DaemonSet":
		pod, err = kk.Ctl().DaemonSet().ManagedPod()
	case "ReplicaSet":
		pod, err = kk.Ctl().ReplicaSet().ManagedPod()
	}
	return pod, err
}
func getPods(selectedCluster string, ctx context.Context, ns string, name string, kind string, group string, version string) ([]*v1.Pod, error) {
	var pods []*v1.Pod
	var err error
	kk := kom.Cluster(selectedCluster).WithContext(ctx).
		CRD(group, version, kind).
		Namespace(ns).
		Name(name).
		WithCache(linkCacheTTL)
	pods, err = kk.Ctl().CRD().ManagedPods()

	if err == nil && len(pods) != 0 {
		return pods, nil
	}
	switch kind {
	case "Deployment":
		pods, err = kk.Ctl().Deployment().ManagedPods()
	case "StatefulSet":
		pods, err = kk.Ctl().StatefulSet().ManagedPods()
	case "DaemonSet":
		pods, err = kk.Ctl().DaemonSet().ManagedPods()
	case "ReplicaSet":
		pods, err = kk.Ctl().ReplicaSet().ManagedPods()
	}
	return pods, err
}

// @Summary 获取Pod关联的服务
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param kind path string true "资源类型"
// @Param group path string true "API组"
// @Param version path string true "API版本"
// @Param ns path string true "命名空间"
// @Param name path string true "资源名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/{kind}/group/{group}/version/{version}/ns/{ns}/name/{name}/links/services [get]
func (pc *PodLinkController) LinksServices(c *response.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	ctx := amis.GetContextWithUser(c)
	kind := c.Param("kind")
	group := c.Param("group")
	version := c.Param("version")
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	pod, err := getPod(selectedCluster, ctx, ns, name, kind, group, version)

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	services, err := service.PodService().LinksServices(ctx, selectedCluster, pod)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonList(c, services)
}

// @Summary 获取Pod关联的端点
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param kind path string true "资源类型"
// @Param group path string true "API组"
// @Param version path string true "API版本"
// @Param ns path string true "命名空间"
// @Param name path string true "资源名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/{kind}/group/{group}/version/{version}/ns/{ns}/name/{name}/links/endpoints [get]
func (pc *PodLinkController) LinksEndpoints(c *response.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	ctx := amis.GetContextWithUser(c)
	kind := c.Param("kind")
	group := c.Param("group")
	version := c.Param("version")
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	pod, err := getPod(selectedCluster, ctx, ns, name, kind, group, version)

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	endpoints, err := service.PodService().LinksEndpoints(ctx, selectedCluster, pod)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonList(c, endpoints)

}

// @Summary 获取Pod关联的PVC
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param kind path string true "资源类型"
// @Param group path string true "API组"
// @Param version path string true "API版本"
// @Param ns path string true "命名空间"
// @Param name path string true "资源名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/{kind}/group/{group}/version/{version}/ns/{ns}/name/{name}/links/pvc [get]
func (pc *PodLinkController) LinksPVC(c *response.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	ctx := amis.GetContextWithUser(c)
	kind := c.Param("kind")
	group := c.Param("group")
	version := c.Param("version")
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	pod, err := getPod(selectedCluster, ctx, ns, name, kind, group, version)

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	pvc, err := service.PodService().LinksPVC(ctx, selectedCluster, pod)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonList(c, pvc)
}

// @Summary 获取Pod关联的PV
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param kind path string true "资源类型"
// @Param group path string true "API组"
// @Param version path string true "API版本"
// @Param ns path string true "命名空间"
// @Param name path string true "资源名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/{kind}/group/{group}/version/{version}/ns/{ns}/name/{name}/links/pv [get]
func (pc *PodLinkController) LinksPV(c *response.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	ctx := amis.GetContextWithUser(c)
	kind := c.Param("kind")
	group := c.Param("group")
	version := c.Param("version")
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	pod, err := getPod(selectedCluster, ctx, ns, name, kind, group, version)

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	pv, err := service.PodService().LinksPV(ctx, selectedCluster, pod)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonList(c, pv)
}

// @Summary 获取Pod关联的Ingress
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param kind path string true "资源类型"
// @Param group path string true "API组"
// @Param version path string true "API版本"
// @Param ns path string true "命名空间"
// @Param name path string true "资源名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/{kind}/group/{group}/version/{version}/ns/{ns}/name/{name}/links/ingress [get]
func (pc *PodLinkController) LinksIngress(c *response.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	ctx := amis.GetContextWithUser(c)
	kind := c.Param("kind")
	group := c.Param("group")
	version := c.Param("version")
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	pod, err := getPod(selectedCluster, ctx, ns, name, kind, group, version)

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	ingress, err := service.PodService().LinksIngress(ctx, selectedCluster, pod)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonList(c, ingress)
}

// @Summary 获取Pod关联的环境变量
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param kind path string true "资源类型"
// @Param group path string true "API组"
// @Param version path string true "API版本"
// @Param ns path string true "命名空间"
// @Param name path string true "资源名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/{kind}/group/{group}/version/{version}/ns/{ns}/name/{name}/links/env [get]
func (pc *PodLinkController) LinksEnv(c *response.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	ctx := amis.GetContextWithUser(c)
	kind := c.Param("kind")
	group := c.Param("group")
	version := c.Param("version")
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	pod, err := getPod(selectedCluster, ctx, ns, name, kind, group, version)

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	env, err := service.PodService().LinksEnv(ctx, selectedCluster, pod)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonList(c, env)
}

// @Summary 获取Pod关联的来自其他Pod的环境变量
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param kind path string true "资源类型"
// @Param group path string true "API组"
// @Param version path string true "API版本"
// @Param ns path string true "命名空间"
// @Param name path string true "资源名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/{kind}/group/{group}/version/{version}/ns/{ns}/name/{name}/links/envFromPod [get]
func (pc *PodLinkController) LinksEnvFromPod(c *response.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	ctx := amis.GetContextWithUser(c)
	kind := c.Param("kind")
	group := c.Param("group")
	version := c.Param("version")
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	pod, err := getPod(selectedCluster, ctx, ns, name, kind, group, version)

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	env, err := service.PodService().LinksEnvFromPod(ctx, selectedCluster, pod)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonList(c, env)
}

// @Summary 获取Pod关联的ConfigMap
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param kind path string true "资源类型"
// @Param group path string true "API组"
// @Param version path string true "API版本"
// @Param ns path string true "命名空间"
// @Param name path string true "资源名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/{kind}/group/{group}/version/{version}/ns/{ns}/name/{name}/links/configmap [get]
func (pc *PodLinkController) LinksConfigMap(c *response.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	ctx := amis.GetContextWithUser(c)
	kind := c.Param("kind")
	group := c.Param("group")
	version := c.Param("version")
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	pod, err := getPod(selectedCluster, ctx, ns, name, kind, group, version)

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	configMap, err := service.PodService().LinksConfigMap(ctx, selectedCluster, pod)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonList(c, configMap)
}

// @Summary 获取Pod关联的Secret
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param kind path string true "资源类型"
// @Param group path string true "API组"
// @Param version path string true "API版本"
// @Param ns path string true "命名空间"
// @Param name path string true "资源名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/{kind}/group/{group}/version/{version}/ns/{ns}/name/{name}/links/secret [get]
func (pc *PodLinkController) LinksSecret(c *response.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	ctx := amis.GetContextWithUser(c)
	kind := c.Param("kind")
	group := c.Param("group")
	version := c.Param("version")
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	pod, err := getPod(selectedCluster, ctx, ns, name, kind, group, version)

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	secret, err := service.PodService().LinksSecret(ctx, selectedCluster, pod)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonList(c, secret)
}

// @Summary 获取Pod关联的节点
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param kind path string true "资源类型"
// @Param group path string true "API组"
// @Param version path string true "API版本"
// @Param ns path string true "命名空间"
// @Param name path string true "资源名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/{kind}/group/{group}/version/{version}/ns/{ns}/name/{name}/links/node [get]
func (pc *PodLinkController) LinksNode(c *response.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	ctx := amis.GetContextWithUser(c)
	kind := c.Param("kind")
	group := c.Param("group")
	version := c.Param("version")
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	pod, err := getPod(selectedCluster, ctx, ns, name, kind, group, version)

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	nodes, err := service.PodService().LinksNode(ctx, selectedCluster, pod)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonList(c, nodes)
}
func (pc *PodLinkController) LinksPod(c *response.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	ctx := amis.GetContextWithUser(c)
	kind := c.Param("kind")
	group := c.Param("group")
	version := c.Param("version")
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	pods, err := getPods(selectedCluster, ctx, ns, name, kind, group, version)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonList(c, pods)
}
