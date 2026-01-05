package dynamic

import (
	"fmt"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/response"
	"github.com/weibaohui/kom/kom"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type MetadataController struct{}

// RegisterMetadataRoutes 注册路由

func RegisterMetadataRoutes(api chi.Router) {
	ctrl := &MetadataController{}
	api.Post("/{kind}/group/{group}/version/{version}/update_labels/ns/{ns}/name/{name}", response.Adapter(ctrl.UpdateLabels))
	api.Get("/{kind}/group/{group}/version/{version}/annotations/ns/{ns}/name/{name}", response.Adapter(ctrl.ListAnnotations))
	api.Post("/{kind}/group/{group}/version/{version}/update_annotations/ns/{ns}/name/{name}", response.Adapter(ctrl.UpdateAnnotations))
}

// @Summary 更新资源标签
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param kind path string true "资源类型"
// @Param group path string true "资源组"
// @Param version path string true "资源版本"
// @Param ns path string true "命名空间"
// @Param name path string true "资源名称"
// @Param labels body map[string]string true "标签键值对"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/{kind}/group/{group}/version/{version}/update_labels/ns/{ns}/name/{name} [post]
func (mc *MetadataController) UpdateLabels(c *response.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	kind := c.Param("kind")
	group := c.Param("group")
	version := c.Param("version")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var req struct {
		Labels map[string]string `json:"labels"`
	}
	if err = c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var obj *unstructured.Unstructured
	err = kom.Cluster(selectedCluster).WithContext(ctx).
		Name(name).Namespace(ns).
		CRD(group, version, kind).
		Get(&obj).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	obj.SetLabels(req.Labels)

	err = kom.Cluster(selectedCluster).WithContext(ctx).
		Name(name).Namespace(ns).
		CRD(group, version, kind).
		Update(obj).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonOK(c)
}

// 部分key为k8m增加的指标数据，不是资源自身的注解，因此过滤掉。
// last-applied-configuration是k8s管理的，不允许修改。
// 注意同步修改前端的assets/public/custom.js里面的filterAnnotations方法
var immutableKeys = []string{
	"cpu.request",
	"cpu.requestFraction",
	"cpu.limit",
	"cpu.limitFraction",
	"cpu.total",
	"cpu.realtime",
	"memory.request",
	"memory.requestFraction",
	"memory.limit",
	"memory.limitFraction",
	"memory.total",
	"memory.realtime",
	"ip.usage.total",
	"ip.usage.used",
	"ip.usage.available",
	"pod.count.total",
	"pod.count.used",
	"pod.count.available",
	"kubectl.kubernetes.io/last-applied-configuration",
	"kom.kubernetes.io/restartedAt",
	"pvc.count",
	"pv.count",
	"ingress.count",
}

// @Summary 列出资源注解
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param kind path string true "资源类型"
// @Param group path string true "资源组"
// @Param version path string true "资源版本"
// @Param ns path string true "命名空间"
// @Param name path string true "资源名称"
// @Success 200 {object} map[string]string
// @Router /k8s/cluster/{cluster}/{kind}/group/{group}/version/{version}/annotations/ns/{ns}/name/{name} [get]
func (mc *MetadataController) ListAnnotations(c *response.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	kind := c.Param("kind")
	group := c.Param("group")
	version := c.Param("version")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var obj *unstructured.Unstructured
	err = kom.Cluster(selectedCluster).WithContext(ctx).
		Name(name).Namespace(ns).
		CRD(group, version, kind).
		Get(&obj).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	annotations := obj.GetAnnotations()
	// 排除immutableKeys
	for _, key := range immutableKeys {
		delete(annotations, key)
	}

	amis.WriteJsonData(c, response.H{
		"annotations": annotations,
	})
}

// @Summary 更新资源注解
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param kind path string true "资源类型"
// @Param group path string true "资源组"
// @Param version path string true "资源版本"
// @Param ns path string true "命名空间"
// @Param name path string true "资源名称"
// @Param annotations body map[string]interface{} true "注解键值对"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/{kind}/group/{group}/version/{version}/update_annotations/ns/{ns}/name/{name} [post]
func (mc *MetadataController) UpdateAnnotations(c *response.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	kind := c.Param("kind")
	group := c.Param("group")
	version := c.Param("version")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var req struct {
		Annotations map[string]any `json:"annotations"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 判断下前台传来的annotations是否是immutableKeys中的key，如果是则不允许修改
	// 创建一个新的map，用于存储过滤后的annotations
	filteredAnnotations := make(map[string]string)

	for k, v := range req.Annotations {
		if !slice.Contain(immutableKeys, k) {
			filteredAnnotations[k] = fmt.Sprintf("%s", v)
		}
	}

	// 判断是否还有值，有值再更新
	if len(filteredAnnotations) == 0 {
		amis.WriteJsonOK(c)
		return
	}
	var obj *unstructured.Unstructured
	err = kom.Cluster(selectedCluster).WithContext(ctx).
		Name(name).Namespace(ns).
		CRD(group, version, kind).
		Get(&obj).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 单独处理kubectl.kubernetes.io/last-applied-configuration
	// 这个要用原来的覆盖
	if obj.GetAnnotations()["kubectl.kubernetes.io/last-applied-configuration"] != "" {
		filteredAnnotations["kubectl.kubernetes.io/last-applied-configuration"] = obj.GetAnnotations()["kubectl.kubernetes.io/last-applied-configuration"]
	}
	obj.SetAnnotations(filteredAnnotations)

	err = kom.Cluster(selectedCluster).WithContext(ctx).
		Name(name).Namespace(ns).
		CRD(group, version, kind).
		Update(obj).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonOK(c)
}
