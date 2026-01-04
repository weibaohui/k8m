package dynamic

import (
	"fmt"

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

type TolerationController struct{}

func RegisterTolerationRoutes(api *chi.Router) {
	ctrl := &TolerationController{}
	api.POST("/{kind}/group/{group}/version/{version}/update_tolerations/ns/{ns}/name/{name}", ctrl.Update)
	api.POST("/{kind}/group/{group}/version/{version}/delete_tolerations/ns/{ns}/name/{name}", ctrl.Delete)
	api.POST("/{kind}/group/{group}/version/{version}/add_tolerations/ns/{ns}/name/{name}", ctrl.Add)
	api.GET("/{kind}/group/{group}/version/{version}/list_tolerations/ns/{ns}/name/{name}", ctrl.List)
}

// @Summary 获取资源容忍度列表
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param kind path string true "资源类型"
// @Param group path string true "API组"
// @Param version path string true "API版本"
// @Param ns path string true "命名空间"
// @Param name path string true "资源名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/{kind}/group/{group}/version/{version}/list_tolerations/ns/{ns}/name/{name} [get]
func (tc *TolerationController) List(c *response.Context) {
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
	tolerationsPath := append(paths, "tolerations")
	// 获取 Affinity 配置并解析
	tolerations, found, err := unstructured.NestedFieldNoCopy(item.Object, tolerationsPath...)
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
	tolerationsList, ok := tolerations.([]any)
	if !ok {
		amis.WriteJsonError(c, fmt.Errorf("nodeSelectorTerms is not an array"))
		return
	}

	amis.WriteJsonList(c, tolerationsList)
}

// @Summary 添加资源容忍度
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param kind path string true "资源类型"
// @Param group path string true "API组"
// @Param version path string true "API版本"
// @Param ns path string true "命名空间"
// @Param name path string true "资源名称"
// @Param body body Tolerations true "容忍度配置信息"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/{kind}/group/{group}/version/{version}/add_tolerations/ns/{ns}/name/{name} [post]
func (tc *TolerationController) Add(c *response.Context) {
	processTolerations(c, "add")
}

// @Summary 更新资源容忍度
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param kind path string true "资源类型"
// @Param group path string true "API组"
// @Param version path string true "API版本"
// @Param ns path string true "命名空间"
// @Param name path string true "资源名称"
// @Param body body Tolerations true "容忍度配置信息"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/{kind}/group/{group}/version/{version}/update_tolerations/ns/{ns}/name/{name} [post]
func (tc *TolerationController) Update(c *response.Context) {
	processTolerations(c, "modify")
}

// @Summary 删除资源容忍度
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param kind path string true "资源类型"
// @Param group path string true "API组"
// @Param version path string true "API版本"
// @Param ns path string true "命名空间"
// @Param name path string true "资源名称"
// @Param body body Tolerations true "容忍度配置信息"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/{kind}/group/{group}/version/{version}/delete_tolerations/ns/{ns}/name/{name} [post]
func (tc *TolerationController) Delete(c *response.Context) {
	processTolerations(c, "delete")
}

func processTolerations(c *response.Context, action string) {
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

	originalTolerations, err := getTolerationList(kind, item, action, info)
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
	klog.V(6).Infof("Update Patch JSON :\n%s\n", patchJSON)
	var obj any
	err = kom.Cluster(selectedCluster).
		WithContext(ctx).
		CRD(group, version, kind).
		Namespace(ns).Name(name).
		Patch(&obj, types.StrategicMergePatchType, patchJSON).Error
	amis.WriteJsonErrorOrOK(c, err)
}

// action : modify\update\add
func getTolerationList(kind string, item *unstructured.Unstructured, action string, rule Tolerations) ([]any, error) {
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
		tolerations = make([]any, 0)
	}

	// 强制转换为数组
	tolerationsList, ok := tolerations.([]any)
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
		tolerationsList = []any{
			map[string]any{
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
	var newTolerationsList []any
	for _, term := range tolerationsList {
		termMap, ok := term.(map[string]any)
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
			newTolerationsList = append(newTolerationsList, map[string]any{
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
		if !slice.ContainBy(tolerationsList, func(item any) bool {
			m := item.(map[string]any)
			return m["key"] == rule.Key && m["value"] == rule.Value && m["effect"] == rule.Effect
		}) {
			newTolerationsList = append(tolerationsList, map[string]any{
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
func generateRequiredTolerationsDynamicPatch(kind string, rules []any) (map[string]any, error) {
	// 获取资源路径
	paths, err := getResourcePaths(kind)
	if err != nil {
		return nil, err
	}

	// 动态构造 patch 数据
	patch := make(map[string]any)
	current := patch

	// 按层级动态生成嵌套结构
	for _, path := range paths {
		if _, exists := current[path]; !exists {
			current[path] = make(map[string]any)
		}
		current = current[path].(map[string]any)
	}

	current["tolerations"] = rules

	return patch, nil
}

type Tolerations struct {
	Operator          string `json:"operator"`
	Key               string `json:"key"`
	Value             string `json:"value"`
	Effect            string `json:"effect"`
	TolerationSeconds *int64 `json:"tolerationSeconds"`
}
