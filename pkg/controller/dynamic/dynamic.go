package dynamic

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

func List(c *gin.Context) {
	ns := c.Param("ns")
	group := c.Param("group")
	kind := c.Param("kind")
	version := c.Param("version")
	ctx := c.Request.Context()
	var list []unstructured.Unstructured
	err := kom.DefaultCluster().WithContext(ctx).RemoveManagedFields().Namespace(ns).GVK(group, version, kind).List(&list).Error
	amis.WriteJsonListWithError(c, list, err)
}
func Event(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	kind := c.Param("kind")
	group := c.Param("group")
	version := c.Param("version")
	ctx := c.Request.Context()

	apiVersion := fmt.Sprintf("%s", version)
	if group != "" {
		apiVersion = fmt.Sprintf("%s/%s", group, version)
	}

	fieldSelector := fmt.Sprintf("regarding.apiVersion=%s,regarding.kind=%s,regarding.name=%s,regarding.namespace=%s", apiVersion, kind, name, ns)

	var eventList []unstructured.Unstructured
	err := kom.DefaultCluster().
		WithContext(ctx).
		RemoveManagedFields().
		Namespace(ns).
		GVK("events.k8s.io", "v1", "Event").
		List(&eventList, metav1.ListOptions{
			FieldSelector: fieldSelector,
		}).Error

	amis.WriteJsonListWithError(c, eventList, err)

}
func Fetch(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	kind := c.Param("kind")
	group := c.Param("group")
	version := c.Param("version")
	ctx := c.Request.Context()

	var obj *unstructured.Unstructured

	err := kom.DefaultCluster().WithContext(ctx).RemoveManagedFields().Name(name).Namespace(ns).CRD(group, version, kind).Get(&obj).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	yamlStr, err := utils.ConvertUnstructuredToYAML(obj)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, gin.H{
		"yaml": yamlStr,
	})
}
func Remove(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	kind := c.Param("kind")
	group := c.Param("group")
	version := c.Param("version")
	ctx := c.Request.Context()

	err := removeSingle(ctx, kind, group, version, ns, name)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)

}
func removeSingle(ctx context.Context, kind, group, version, ns, name string) error {
	return kom.DefaultCluster().WithContext(ctx).Name(name).Namespace(ns).CRD(group, version, kind).Delete().Error
}

// NamesPayload 定义结构体以匹配批量删除 JSON 结构
type NamesPayload struct {
	Names []string `json:"names"`
}

func BatchRemove(c *gin.Context) {
	ns := c.Param("ns")
	kind := c.Param("kind")
	group := c.Param("group")
	version := c.Param("version")
	ctx := c.Request.Context()

	// 初始化结构体实例
	var payload NamesPayload

	// 反序列化 JSON 数据到结构体
	if err := c.ShouldBindJSON(&payload); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	for _, name := range payload.Names {
		_ = removeSingle(ctx, kind, group, version, ns, name)
	}
	amis.WriteJsonOK(c)
}

type ApplyYAMLRequest struct {
	YAML string `json:"yaml" binding:"required"`
}

func Save(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	kind := c.Param("kind")
	group := c.Param("group")
	version := c.Param("version")
	ctx := c.Request.Context()

	var req ApplyYAMLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	yamlStr := req.YAML

	// 解析 YAML 到 Unstructured 对象
	var obj unstructured.Unstructured
	if err := yaml.Unmarshal([]byte(yamlStr), &obj.Object); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	obj.SetName(name)
	obj.SetNamespace(ns)
	err := kom.DefaultCluster().WithContext(ctx).Name(name).Namespace(ns).CRD(group, version, kind).Update(&obj).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

func Describe(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	kind := c.Param("kind")
	group := c.Param("group")
	version := c.Param("version")
	ctx := c.Request.Context()
	var result []byte

	err := kom.DefaultCluster().WithContext(ctx).Name(name).Namespace(ns).CRD(group, version, kind).Describe(&result).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, string(result))
}

func Apply(c *gin.Context) {
	ctx := c.Request.Context()

	var req ApplyYAMLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	yamlStr := req.YAML
	result := kom.DefaultCluster().WithContext(ctx).Applier().Apply(yamlStr)
	amis.WriteJsonData(c, gin.H{
		"result": result,
	})

}
func Delete(c *gin.Context) {
	ctx := c.Request.Context()

	var req ApplyYAMLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	yamlStr := req.YAML
	result := kom.DefaultCluster().WithContext(ctx).Applier().Delete(yamlStr)
	amis.WriteJsonData(c, gin.H{
		"result": result,
	})
}
