package dynamic

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

// {"labelSelector":{"matchLabels":{"app":"other-app","x":"y"}},"topologyKey":"kubernetes.io/hostname"}
type podAffinity struct {
	LabelSelector struct {
		MatchLabels map[string]string `json:"matchLabels"`
	} `json:"labelSelector"`
	TopologyKey string `json:"topologyKey"`
}

// ListPodAffinity 获取 Pod 亲和性列表
func ListPodAffinity(c *gin.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	group := c.Param("group")
	kind := c.Param("kind")
	version := c.Param("version")
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	// 先获取资源中的定义
	var item unstructured.Unstructured
	err := kom.Cluster(selectedCluster).
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
	requiredAffinityPath := append(paths, "affinity", "podAffinity", "requiredDuringSchedulingIgnoredDuringExecution")
	// 获取 Affinity 配置并解析
	affinity, found, err := unstructured.NestedFieldNoCopy(item.Object, requiredAffinityPath...)
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
	list, ok := affinity.([]interface{})
	if !ok {
		amis.WriteJsonError(c, fmt.Errorf("list is not an array"))
		return
	}
	amis.WriteJsonList(c, list)
}

func AddPodAffinity(c *gin.Context) {
	processPodAffinity(c, "add")
}
func UpdatePodAffinity(c *gin.Context) {
	processPodAffinity(c, "modify")
}
func DeletePodAffinity(c *gin.Context) {
	processPodAffinity(c, "delete")
}
func processPodAffinity(c *gin.Context, action string) {
	name := c.Param("name")
	ns := c.Param("ns")
	group := c.Param("group")
	kind := c.Param("kind")
	version := c.Param("version")
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	var info nodeAffinity

	if err := c.ShouldBindJSON(&info); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 强制性规则处理。
	// operator: In 时，values必须有值。
	// operator: Exists DoesNotExist 时，values必须没有值。
	if info.Operator == "Exists" || info.Operator == "DoesNotExist" {
		info.Values = make([]string, 0)
	}

	// 先获取资源中的定义
	var item unstructured.Unstructured
	err := kom.Cluster(selectedCluster).
		WithContext(ctx).
		CRD(group, version, kind).
		Namespace(ns).Name(name).
		Get(&item).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	originalNodeAffinity, err := getNodeSelectorTerms(kind, &item, action, info)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	patchData, err := generateRequiredNodeAffinityDynamicPatch(kind, originalNodeAffinity)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	patchJSON := utils.ToJSON(patchData)
	klog.V(6).Infof("UpdateNodeAffinity Patch JSON :\n%s\n", patchJSON)
	var obj interface{}
	err = kom.Cluster(selectedCluster).
		WithContext(ctx).
		CRD(group, version, kind).
		Namespace(ns).Name(name).
		Patch(&obj, types.MergePatchType, patchJSON).Error
	amis.WriteJsonErrorOrOK(c, err)
}

// getNodeSelectorTerms 获取 NodeAffinity 的 NodeSelectorTerms
// action : modify\update\add
func getPodSelectorTerms(kind string, item *unstructured.Unstructured, action string, rule nodeAffinity) ([]interface{}, error) {

	// 获取资源路径
	paths, err := getResourcePaths(kind)
	if err != nil {
		return nil, err
	}
	requiredAffinityPath := append(paths, "affinity", "nodeAffinity", "requiredDuringSchedulingIgnoredDuringExecution", "nodeSelectorTerms")

	// 获取 Affinity 配置并解析
	affinity, found, err := unstructured.NestedFieldNoCopy(item.Object, requiredAffinityPath...)
	if err != nil {
		return nil, err
	}
	if !found {
		// 如果没有，那么如果是 add 操作，那么创建一个空的 NodeSelectorTerms
		affinity = make([]interface{}, 0)
	}

	// 强制转换为数组
	nodeSelectorTerms, ok := affinity.([]interface{})
	if !ok {
		return nil, fmt.Errorf("nodeSelectorTerms is not an array")
	}
	x := utils.ToJSON(affinity)
	// 如果nodeSelectorTerms被设置为-{},那么json输出为
	// [
	//  {}
	// ]
	// 其长度为8
	if (len(nodeSelectorTerms) == 0 || len(x) == 8) && action == "add" {
		nodeSelectorTerms = []interface{}{
			map[string]interface{}{
				"matchExpressions": []map[string]interface{}{{
					"key":      rule.Key,
					"operator": rule.Operator,
					"values":   rule.Values,
				},
				},
			},
		}
		return nodeSelectorTerms, nil
	}

	// 进行操作：删除或新增
	var newNodeSelectorTerms []interface{}
	for _, term := range nodeSelectorTerms {
		termMap, ok := term.(map[string]interface{})
		if !ok {
			continue
		}

		// 获取 matchExpressions
		matchExpressions, found, err := unstructured.NestedFieldNoCopy(termMap, "matchExpressions")
		if err != nil {
			return nil, err
		}
		if !found {
			continue
		}

		// 强制转换为数组
		matchExpressionsArray, ok := matchExpressions.([]interface{})
		if !ok {
			continue
		}

		// 处理每一个 matchExpression
		var newMatchExpressions []interface{}
		for _, expr := range matchExpressionsArray {
			exprMap, ok := expr.(map[string]interface{})
			if !ok {
				continue
			}

			// 比对删除的 key
			if action == "delete" && exprMap["key"] == rule.Key {
				// 如果是删除，并且 key 匹配，跳过这个 matchExpression
				continue
			}
			// 如果是修改操作，并且 key 匹配，修改该 matchExpression 的 values
			if action == "modify" && exprMap["key"] == rule.Key {
				// 修改新值，不修改原值，直接添加到新列表
				newMatchExpressions = append(newMatchExpressions, map[string]interface{}{
					"key":      rule.Key,
					"operator": rule.Operator,
					"values":   rule.Values,
				})
				continue
			}
			// 否则，保留原有的 matchExpression
			newMatchExpressions = append(newMatchExpressions, map[string]interface{}{
				"key":      exprMap["key"],
				"operator": exprMap["operator"],
				"values":   exprMap["values"],
			})
		}

		// 如果是新增操作，增加新的 matchExpression
		if action == "add" {
			newMatchExpressions = append(newMatchExpressions, map[string]interface{}{
				"key":      rule.Key,
				"operator": rule.Operator,
				"values":   rule.Values,
			})
		}

		// 将修改后的 matchExpressions 赋值回 termMap
		// 这里直接将新的 matchExpressions 更新到 termMap 中，不使用 SetNestedField
		termMap["matchExpressions"] = newMatchExpressions

		// 将修改后的 termMap 添加到新的 nodeSelectorTerms 列表中
		newNodeSelectorTerms = append(newNodeSelectorTerms, termMap)
	}

	return newNodeSelectorTerms, nil
}

// 生成动态的 patch 数据
func generateRequiredPodAffinityDynamicPatch(kind string, rules []interface{}) (map[string]interface{}, error) {
	// 获取资源路径
	paths, err := getResourcePaths(kind)
	if err != nil {
		return nil, err
	}
	requiredAffinityPath := append(paths, "affinity", "nodeAffinity", "requiredDuringSchedulingIgnoredDuringExecution")

	// 动态构造 patch 数据
	patch := make(map[string]interface{})
	current := patch

	// 按层级动态生成嵌套结构
	for _, path := range requiredAffinityPath {
		if _, exists := current[path]; !exists {
			current[path] = make(map[string]interface{})
		}
		current = current[path].(map[string]interface{})
	}

	current["nodeSelectorTerms"] = rules

	return patch, nil
}
