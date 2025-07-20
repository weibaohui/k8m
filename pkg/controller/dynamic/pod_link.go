package dynamic

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
)

type PodLinkController struct{}

func RegisterPodLinkRoutes(api *gin.RouterGroup) {
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
func (pc *PodLinkController) LinksServices(c *gin.Context) {
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

func (pc *PodLinkController) LinksEndpoints(c *gin.Context) {
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

func (pc *PodLinkController) LinksPVC(c *gin.Context) {
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

func (pc *PodLinkController) LinksPV(c *gin.Context) {
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

func (pc *PodLinkController) LinksIngress(c *gin.Context) {
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

func (pc *PodLinkController) LinksEnv(c *gin.Context) {
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

func (pc *PodLinkController) LinksEnvFromPod(c *gin.Context) {
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

func (pc *PodLinkController) LinksConfigMap(c *gin.Context) {
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

func (pc *PodLinkController) LinksSecret(c *gin.Context) {
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

func (pc *PodLinkController) LinksNode(c *gin.Context) {
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
func (pc *PodLinkController) LinksPod(c *gin.Context) {
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
