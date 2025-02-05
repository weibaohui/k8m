package pod

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
)

var linkCacheTTL = 5 * time.Minute

func LinksServices(c *gin.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)
	var pod *v1.Pod
	err := kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.Pod{}).
		Namespace(ns).
		Name(name).
		WithCache(linkCacheTTL).
		Get(&pod).Error
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

func LinksEndpoints(c *gin.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)
	var pod *v1.Pod
	err := kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.Pod{}).
		Namespace(ns).
		Name(name).
		WithCache(linkCacheTTL).
		Get(&pod).Error
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

func LinksPVC(c *gin.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)
	var pod *v1.Pod
	err := kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.Pod{}).
		Namespace(ns).
		Name(name).
		WithCache(linkCacheTTL).
		Get(&pod).Error
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

func LinksPV(c *gin.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)
	var pod *v1.Pod
	err := kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.Pod{}).
		Namespace(ns).
		Name(name).
		WithCache(linkCacheTTL).
		Get(&pod).Error
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

func LinksIngress(c *gin.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)
	var pod *v1.Pod
	err := kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.Pod{}).
		Namespace(ns).
		Name(name).
		WithCache(linkCacheTTL).
		Get(&pod).Error
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

func LinksEnv(c *gin.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)
	var pod *v1.Pod
	err := kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.Pod{}).
		Namespace(ns).
		Name(name).
		WithCache(linkCacheTTL).
		Get(&pod).Error
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

func LinksEnvFromPod(c *gin.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)
	var pod *v1.Pod
	err := kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.Pod{}).
		Namespace(ns).
		Name(name).
		WithCache(linkCacheTTL).
		Get(&pod).Error
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

func LinksConfigMap(c *gin.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)
	var pod *v1.Pod
	err := kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.Pod{}).
		Namespace(ns).
		Name(name).
		WithCache(linkCacheTTL).
		Get(&pod).Error
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

func LinksSecret(c *gin.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)
	var pod *v1.Pod
	err := kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.Pod{}).
		Namespace(ns).
		Name(name).
		WithCache(linkCacheTTL).
		Get(&pod).Error
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

func LinksNode(c *gin.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)
	var pod *v1.Pod
	err := kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.Pod{}).
		Namespace(ns).
		Name(name).
		WithCache(linkCacheTTL).
		Get(&pod).Error
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
