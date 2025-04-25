package dynamic

import (
	"fmt"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

type Tolerations struct {
	Operator          string `json:"operator"`
	Key               string `json:"key"`
	Value             string `json:"value"`
	Effect            string `json:"effect"`
	TolerationSeconds *int64 `json:"tolerationSeconds"`
}

func ListTolerations(c *gin.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	group := c.Param("group")
	kind := c.Param("kind")
	version := c.Param("version")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 先获取资源中的定义
	var item unstructured.Unstructured
	err = kom.Cluster(selectedCluster).
		WithContext(ctx).
		CRD(group, version, kind).
		Namespace(ns).Name(name).
		Get(&item).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 获取资源路径
	paths, err := getResourcePaths(kind)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	tolerationsPath := append(paths, "tolerations")
	// 获取 Affinity 配置并解析
	tolerations, found, err := unstructured.NestedFieldNoCopy(item.Object, tolerationsPath...)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	if !found {
		// 没有的话，返回一个空列表
		amis.WriteJsonList(c, []interface{}{})
		return
	}

	// 强制转换为数组
	tolerationsList, ok := tolerations.([]interface{})
	if !ok {
		amis.WriteJsonError(c, fmt.Errorf("nodeSelectorTerms is not an array"))
		return
	}

	amis.WriteJsonList(c, tolerationsList)
}
func AddTolerations(c *gin.Context) {
	processTolerations(c, "add")
}
func UpdateTolerations(c *gin.Context) {
	processTolerations(c, "modify")
}
func DeleteTolerations(c *gin.Context) {
	processTolerations(c, "delete")
}
func processTolerations(c *gin.Context, action string) {
	name := c.Param("name")
	ns := c.Param("ns")
	group := c.Param("group")
	kind := c.Param("kind")
	version := c.Param("version")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var info Tolerations

	if err := c.ShouldBindJSON(&info); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 如果operator 是存在，则不需要设置value值
	if info.Operator == "Exists" {
		info.Value = ""
	}

	// 先获取资源中的定义
	var item unstructured.Unstructured
	err = kom.Cluster(selectedCluster).
		WithContext(ctx).
		CRD(group, version, kind).
		Namespace(ns).Name(name).
		Get(&item).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	originalTolerations, err := getTolerationList(kind, &item, action, info)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	patchData, err := generateRequiredTolerationsDynamicPatch(kind, originalTolerations)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	patchJSON := utils.ToJSON(patchData)
	klog.V(6).Infof("UpdateTolerations Patch JSON :\n%s\n", patchJSON)
	var obj interface{}
	err = kom.Cluster(selectedCluster).
		WithContext(ctx).
		CRD(group, version, kind).
		Namespace(ns).Name(name).
		Patch(&obj, types.StrategicMergePatchType, patchJSON).Error
	amis.WriteJsonErrorOrOK(c, err)
}

// action : modify\update\add
func getTolerationList(kind string, item *unstructured.Unstructured, action string, rule Tolerations) ([]interface{}, error) {
	// 获取资源路径
	paths, err := getResourcePaths(kind)
	if err != nil {
		return nil, err
	}
	tolerationsPath := append(paths, "tolerations")
	// 获取 Affinity 配置并解析
	tolerations, found, err := unstructured.NestedFieldNoCopy(item.Object, tolerationsPath...)
	if err != nil {
		return nil, err
	}
	if !found {
		// 没有的话，返回一个空列表
		tolerations = make([]interface{}, 0)
	}

	// 强制转换为数组
	tolerationsList, ok := tolerations.([]interface{})
	if !ok {
		return nil, fmt.Errorf("tolerations is not an array")

	}

	x := utils.ToJSON(tolerations)
	// 如果nodeSelectorTerms被设置为-{},那么json输出为
	// [
	//  {}
	// ]
	// 其长度为8
	if (len(tolerationsList) == 0 || len(x) == 8) && action == "add" {
		tolerationsList = []interface{}{
			map[string]interface{}{
				"key":               rule.Key,
				"operator":          rule.Operator,
				"value":             rule.Value,
				"effect":            rule.Effect,
				"tolerationSeconds": rule.TolerationSeconds,
			},
		}
		return tolerationsList, nil
	}

	// 进行操作：删除或新增
	var newTolerationsList []interface{}
	for _, term := range tolerationsList {
		termMap, ok := term.(map[string]interface{})
		if !ok {
			continue
		}
		// 比对删除的 key
		if action == "delete" && termMap["key"] == rule.Key {
			// 如果是删除，并且 key 匹配，跳过这个 matchExpression
			continue
		}
		// 如果是修改操作，并且 key 匹配， 那么修改，如果不是，那么直接添加
		if action == "modify" && termMap["key"] == rule.Key {
			newTolerationsList = append(newTolerationsList, map[string]interface{}{
				"key":               rule.Key,
				"operator":          rule.Operator,
				"value":             rule.Value,
				"effect":            rule.Effect,
				"tolerationSeconds": rule.TolerationSeconds,
			})
		} else {
			newTolerationsList = append(newTolerationsList, term)
		}
	}
	// 如果是新增操作，增加
	if action == "add" {
		// 需要判断下是否已经存在
		if !slice.ContainBy(tolerationsList, func(item interface{}) bool {
			m := item.(map[string]interface{})
			return m["key"] == rule.Key && m["value"] == rule.Value && m["effect"] == rule.Effect
		}) {
			newTolerationsList = append(tolerationsList, map[string]interface{}{
				"key":               rule.Key,
				"operator":          rule.Operator,
				"value":             rule.Value,
				"effect":            rule.Effect,
				"tolerationSeconds": rule.TolerationSeconds,
			})
		}

	}
	return newTolerationsList, nil
}

// 生成动态的 patch 数据
func generateRequiredTolerationsDynamicPatch(kind string, rules []interface{}) (map[string]interface{}, error) {
	// 获取资源路径
	paths, err := getResourcePaths(kind)
	if err != nil {
		return nil, err
	}

	// 动态构造 patch 数据
	patch := make(map[string]interface{})
	current := patch

	// 按层级动态生成嵌套结构
	for _, path := range paths {
		if _, exists := current[path]; !exists {
			current[path] = make(map[string]interface{})
		}
		current = current[path].(map[string]interface{})
	}

	current["tolerations"] = rules

	return patch, nil
}
