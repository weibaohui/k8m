package dynamic

import (
	"fmt"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

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

func UpdateAnnotations(c *gin.Context) {
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
		Annotations map[string]interface{} `json:"annotations"`
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
func ListAnnotations(c *gin.Context) {
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

	amis.WriteJsonData(c, gin.H{
		"annotations": annotations,
	})
}
