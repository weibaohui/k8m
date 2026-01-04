package dynamic

import (
	"fmt"
	"strings"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/response"
	"github.com/weibaohui/kom/kom"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

type NodeAffinityController struct{}

// RegisterNodeAffinityRoutes 注册路由
// 从 gin 切换到 chi，使用 chi.Router 替代 gin.RouterGroup
func RegisterNodeAffinityRoutes(api chi.Router) {
	ctrl := &NodeAffinityController{}
	api.Post("/{kind}/group/{group}/version/{version}/update_node_affinity/ns/{ns}/name/{name}", response.Adapter(ctrl.UpdateNodeAffinity))
	api.Post("/{kind}/group/{group}/version/{version}/delete_node_affinity/ns/{ns}/name/{name}", response.Adapter(ctrl.DeleteNodeAffinity))
	api.Post("/{kind}/group/{group}/version/{version}/add_node_affinity/ns/{ns}/name/{name}", response.Adapter(ctrl.AddNodeAffinity))
	api.Get("/{kind}/group/{group}/version/{version}/list_node_affinity/ns/{ns}/name/{name}", response.Adapter(ctrl.ListNodeAffinity))

}

type nodeAffinity struct {
	Operator string   `json:"operator"`
	Key      string   `json:"key"`
	Values   []string `json:"values"`
}

// @Summary 获取节点亲和性配置
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param kind path string true "资源类型"
// @Param group path string true "资源组"
// @Param version path string true "资源版本"
// @Param ns path string true "命名空间"
// @Param name path string true "资源名称"
// @Success 200 {array} interface{}
// @Router /k8s/cluster/{cluster}/{kind}/group/{group}/version/{version}/list_node_affinity/ns/{ns}/name/{name} [get]
func (ac *NodeAffinityController) ListNodeAffinity(c *response.Context) {
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
	var item *unstructured.Unstructured
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
	requiredAffinityPath := append(paths, "affinity", "nodeAffinity", "requiredDuringSchedulingIgnoredDuringExecution", "nodeSelectorTerms")
	// 获取 Affinity 配置并解析
	affinity, found, err := unstructured.NestedFieldNoCopy(item.Object, requiredAffinityPath...)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	if !found {
		// 没有的话，返回一个空列表
		amis.WriteJsonList(c, []any{})
		return
	}

	// 强制转换为数组
	nodeSelectorTerms, ok := affinity.([]any)
	if !ok {
		amis.WriteJsonError(c, fmt.Errorf("nodeSelectorTerms is not an array"))
		return
	}

	var matchExpressionsList []any
	// 遍历 nodeSelectorTerms 来提取 matchExpressions
	for _, term := range nodeSelectorTerms {
		termMap, ok := term.(map[string]any)
		if !ok {
			continue
		}

		// 获取 matchExpressions
		matchExpressions, found, err := unstructured.NestedFieldNoCopy(termMap, "matchExpressions")
		if err != nil {
			amis.WriteJsonError(c, err)
			return
		}
		if !found {
			continue
		}

		// 强制转换为数组，并将其添加到 matchExpressionsList
		matchExpressionsArray, ok := matchExpressions.([]any)
		if ok {
			matchExpressionsList = append(matchExpressionsList, matchExpressionsArray...)
		}
	}

	amis.WriteJsonList(c, matchExpressionsList)
}

// @Summary 添加节点亲和性配置
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param kind path string true "资源类型"
// @Param group path string true "资源组"
// @Param version path string true "资源版本"
// @Param ns path string true "命名空间"
// @Param name path string true "资源名称"
// @Param nodeAffinity body nodeAffinity true "节点亲和性配置"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/{kind}/group/{group}/version/{version}/add_node_affinity/ns/{ns}/name/{name} [post]
func (ac *NodeAffinityController) AddNodeAffinity(c *response.Context) {
	processNodeAffinity(c, "add")
}

// @Summary 更新节点亲和性配置
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param kind path string true "资源类型"
// @Param group path string true "资源组"
// @Param version path string true "资源版本"
// @Param ns path string true "命名空间"
// @Param name path string true "资源名称"
// @Param nodeAffinity body nodeAffinity true "节点亲和性配置"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/{kind}/group/{group}/version/{version}/update_node_affinity/ns/{ns}/name/{name} [post]
func (ac *NodeAffinityController) UpdateNodeAffinity(c *response.Context) {
	processNodeAffinity(c, "modify")
}

// @Summary 删除节点亲和性配置
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param kind path string true "资源类型"
// @Param group path string true "资源组"
// @Param version path string true "资源版本"
// @Param ns path string true "命名空间"
// @Param name path string true "资源名称"
// @Param nodeAffinity body nodeAffinity true "节点亲和性配置"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/{kind}/group/{group}/version/{version}/delete_node_affinity/ns/{ns}/name/{name} [post]
func (ac *NodeAffinityController) DeleteNodeAffinity(c *response.Context) {
	processNodeAffinity(c, "delete")
}
func processNodeAffinity(c *response.Context, action string) {
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

	var info nodeAffinity

	if err = c.ShouldBindJSON(&info); err != nil {
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
	var item *unstructured.Unstructured
	err = kom.Cluster(selectedCluster).
		WithContext(ctx).
		CRD(group, version, kind).
		Namespace(ns).Name(name).
		Get(&item).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	originalNodeAffinity, err := getNodeSelectorTerms(kind, item, action, info)
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
	// 将{}替换为null，为null才会删除
	patchJSON = strings.ReplaceAll(patchJSON, "{}", "null")
	klog.V(6).Infof("UpdateNodeAffinity Patch JSON :\n%s\n", patchJSON)
	var obj any
	err = kom.Cluster(selectedCluster).
		WithContext(ctx).
		CRD(group, version, kind).
		Namespace(ns).Name(name).
		Patch(&obj, types.StrategicMergePatchType, patchJSON).Error
	amis.WriteJsonErrorOrOK(c, err)
}

// getNodeSelectorTerms 获取 NodeAffinity 的 NodeSelectorTerms
// action : modify\update\add
func getNodeSelectorTerms(kind string, item *unstructured.Unstructured, action string, rule nodeAffinity) ([]any, error) {

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
		affinity = make([]any, 0)
	}

	// 强制转换为数组
	nodeSelectorTerms, ok := affinity.([]any)
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
		nodeSelectorTerms = []any{
			map[string]any{
				"matchExpressions": []map[string]any{{
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
	var newNodeSelectorTerms []any
	for _, term := range nodeSelectorTerms {
		termMap, ok := term.(map[string]any)
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
		matchExpressionsArray, ok := matchExpressions.([]any)
		if !ok {
			continue
		}

		// 处理每一个 matchExpression
		var newMatchExpressions []any
		for _, expr := range matchExpressionsArray {
			exprMap, ok := expr.(map[string]any)
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
				newMatchExpressions = append(newMatchExpressions, map[string]any{
					"key":      rule.Key,
					"operator": rule.Operator,
					"values":   rule.Values,
				})
				continue
			}
			// 否则，保留原有的 matchExpression
			newMatchExpressions = append(newMatchExpressions, map[string]any{
				"key":      exprMap["key"],
				"operator": exprMap["operator"],
				"values":   exprMap["values"],
			})
		}

		// 如果是新增操作，增加新的 matchExpression
		if action == "add" {
			if !slice.ContainBy(newMatchExpressions, func(item any) bool {
				m := item.(map[string]any)
				return m["key"] == rule.Key
			}) {
				newMatchExpressions = append(newMatchExpressions, map[string]any{
					"key":      rule.Key,
					"operator": rule.Operator,
					"values":   rule.Values,
				})
			}
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
func generateRequiredNodeAffinityDynamicPatch(kind string, rules []any) (map[string]any, error) {
	// 打印rules
	klog.V(6).Infof("generateRequiredNodeAffinityDynamicPatch rules:\n%+v len=%d\n", rules, len(rules))

	// 删除时[map[matchExpressions:[]]] len=1
	// 先判断是不是要删除,删除层级不一样，单独处理
	if utils.ToJSON(rules) == `[
  {
    "matchExpressions": null
  }
]` {
		return makeDeleteNodeAffinityDynamicPatch(kind, rules)
	}

	// 获取资源路径
	paths, err := getResourcePaths(kind)
	if err != nil {
		return nil, err
	}
	requiredAffinityPath := append(paths, "affinity", "nodeAffinity", "requiredDuringSchedulingIgnoredDuringExecution")

	// 动态构造 patch 数据
	patch := make(map[string]any)
	current := patch

	// 按层级动态生成嵌套结构
	for _, path := range requiredAffinityPath {
		if _, exists := current[path]; !exists {
			current[path] = make(map[string]any)
		}
		current = current[path].(map[string]any)
	}

	current["nodeSelectorTerms"] = rules

	return patch, nil
}
func makeDeleteNodeAffinityDynamicPatch(kind string, rules []any) (map[string]any, error) {
	// 获取资源路径
	paths, err := getResourcePaths(kind)
	if err != nil {
		return nil, err
	}
	requiredAffinityPath := append(paths, "affinity", "nodeAffinity")

	// 动态构造 patch 数据
	patch := make(map[string]any)
	current := patch

	// 按层级动态生成嵌套结构
	for _, path := range requiredAffinityPath {
		if _, exists := current[path]; !exists {
			current[path] = make(map[string]any)
		}
		current = current[path].(map[string]any)
	}

	current = nil
	return patch, nil
}
