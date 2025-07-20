package dynamic

import (
	"fmt"
	"strings"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

type PodAffinityController struct{}

func RegisterPodAffinityRoutes(api *gin.RouterGroup) {
	ctrl := &PodAffinityController{}
	api.POST("/:kind/group/:group/version/:version/update_pod_affinity/ns/:ns/name/:name", ctrl.UpdatePodAffinity)
	api.POST("/:kind/group/:group/version/:version/delete_pod_affinity/ns/:ns/name/:name", ctrl.DeletePodAffinity)
	api.POST("/:kind/group/:group/version/:version/add_pod_affinity/ns/:ns/name/:name", ctrl.AddPodAffinity)
	api.GET("/:kind/group/:group/version/:version/list_pod_affinity/ns/:ns/name/:name", ctrl.ListPodAffinity)

}

// {"labelSelector":{"matchLabels":{"app":"other-app","x":"y"}},"topologyKey":"kubernetes.io/hostname"}
type podAffinity struct {
	LabelSelector struct {
		MatchLabels map[string]string `json:"matchLabels"`
	} `json:"labelSelector"`
	TopologyKey string `json:"topologyKey"`
}

// ListPodAffinity 获取 Pod 亲和性列表
// @Summary 获取Pod亲和性配置
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param kind path string true "资源类型"
// @Param group path string true "资源组"
// @Param version path string true "资源版本"
// @Param ns path string true "命名空间"
// @Param name path string true "资源名称"
// @Success 200 {array} interface{}
// @Router /k8s/cluster/{cluster}/{kind}/group/{group}/version/{version}/list_pod_affinity/ns/{ns}/name/{name} [get]
func (ac *PodAffinityController) ListPodAffinity(c *gin.Context) {
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

// @Summary 添加Pod亲和性配置
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param kind path string true "资源类型"
// @Param group path string true "资源组"
// @Param version path string true "资源版本"
// @Param ns path string true "命名空间"
// @Param name path string true "资源名称"
// @Param podAffinity body podAffinity true "Pod亲和性配置"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/{kind}/group/{group}/version/{version}/add_pod_affinity/ns/{ns}/name/{name} [post]
func (ac *PodAffinityController) AddPodAffinity(c *gin.Context) {
	processPodAffinity(c, "add")
}

// @Summary 更新Pod亲和性配置
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param kind path string true "资源类型"
// @Param group path string true "资源组"
// @Param version path string true "资源版本"
// @Param ns path string true "命名空间"
// @Param name path string true "资源名称"
// @Param podAffinity body podAffinity true "Pod亲和性配置"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/{kind}/group/{group}/version/{version}/update_pod_affinity/ns/{ns}/name/{name} [post]
func (ac *PodAffinityController) UpdatePodAffinity(c *gin.Context) {
	processPodAffinity(c, "modify")
}

// @Summary 删除Pod亲和性配置
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param kind path string true "资源类型"
// @Param group path string true "资源组"
// @Param version path string true "资源版本"
// @Param ns path string true "命名空间"
// @Param name path string true "资源名称"
// @Param podAffinity body podAffinity true "Pod亲和性配置"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/{kind}/group/{group}/version/{version}/delete_pod_affinity/ns/{ns}/name/{name} [post]
func (ac *PodAffinityController) DeletePodAffinity(c *gin.Context) {
	processPodAffinity(c, "delete")
}
func processPodAffinity(c *gin.Context, action string) {
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

	var info podAffinity

	if err = c.ShouldBindJSON(&info); err != nil {
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

	originalPodAffinity, err := getPodAffinityTerms(kind, item, action, info)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	patchData, err := generateRequiredPodAffinityDynamicPatch(kind, originalPodAffinity)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	patchJSON := utils.ToJSON(patchData)
	// 将{}替换为null，为null才会删除
	patchJSON = strings.ReplaceAll(patchJSON, "{}", "null")
	klog.V(6).Infof("UpdatePodAffinity Patch JSON :\n%s\n", patchJSON)
	var obj interface{}
	err = kom.Cluster(selectedCluster).
		WithContext(ctx).
		CRD(group, version, kind).
		Namespace(ns).Name(name).
		Patch(&obj, types.StrategicMergePatchType, patchJSON).Error
	amis.WriteJsonErrorOrOK(c, err)
}

// getPodAffinityTerms 获取 PodAffinity 的 Affinity Terms
// action : modify\update\add
func getPodAffinityTerms(kind string, item *unstructured.Unstructured, action string, rule podAffinity) ([]interface{}, error) {

	// 获取资源路径
	paths, err := getResourcePaths(kind)
	if err != nil {
		return nil, err
	}
	requiredAffinityPath := append(paths, "affinity", "podAffinity", "requiredDuringSchedulingIgnoredDuringExecution")

	// 获取 Affinity 配置并解析
	affinity, found, err := unstructured.NestedFieldNoCopy(item.Object, requiredAffinityPath...)
	if err != nil {
		return nil, err
	}
	if !found {
		// 如果没有，那么如果是 add 操作，那么创建一个空的 podAffinityTerms
		affinity = make([]interface{}, 0)
	}

	// 强制转换为数组
	podAffinityTerms, ok := affinity.([]interface{})
	if !ok {
		return nil, fmt.Errorf("podAffinityTerms is not an array")
	}

	x := utils.ToJSON(affinity)
	// 如果nodeSelectorTerms被设置为-{},那么json输出为
	// [
	//  {}
	// ]
	// 其长度为8
	if (len(podAffinityTerms) == 0 || len(x) == 8) && action == "add" {
		podAffinityTerms = []interface{}{
			map[string]interface{}{
				"labelSelector": map[string]interface{}{
					"matchLabels": rule.LabelSelector.MatchLabels,
				},
				"topologyKey": rule.TopologyKey,
			},
		}
		return podAffinityTerms, nil
	}

	// 进行操作：删除或新增
	var newPodAffinityTerms []interface{}
	for _, term := range podAffinityTerms {
		termMap, ok := term.(map[string]interface{})
		if !ok {
			continue
		}

		// 获取 topologyKey
		topologyKey, found, err := unstructured.NestedFieldNoCopy(termMap, "topologyKey")
		if err != nil {
			return nil, err
		}
		if !found {
			continue
		}

		// 处理每一个 matchLabels
		if action == "delete" {
			if topologyKey == rule.TopologyKey {
				// 如果是删除，并且 matchLabels 和 topologyKey 匹配，跳过这个 term
				continue
			}
		}

		// 如果是修改操作，并且 matchLabels 和 topologyKey 匹配，修改该 term 的 matchLabels
		if action == "modify" && topologyKey == rule.TopologyKey {
			// 修改新值，不修改原值，直接添加到新列表
			newPodAffinityTerms = append(newPodAffinityTerms, map[string]interface{}{
				"labelSelector": map[string]interface{}{
					"matchLabels": rule.LabelSelector.MatchLabels,
				},
				"topologyKey": rule.TopologyKey,
			})
			continue
		}

		// 否则，保留原有的 term
		newPodAffinityTerms = append(newPodAffinityTerms, termMap)
	}

	// 如果是新增操作，增加新的 podAffinityTerm
	if action == "add" {
		if !slice.ContainBy(newPodAffinityTerms, func(item interface{}) bool {
			m := item.(map[string]interface{})
			return m["topologyKey"] == rule.TopologyKey
		}) {
			newPodAffinityTerms = append(newPodAffinityTerms, map[string]interface{}{
				"labelSelector": map[string]interface{}{
					"matchLabels": rule.LabelSelector.MatchLabels,
				},
				"topologyKey": rule.TopologyKey,
			})
		}
	}

	return newPodAffinityTerms, nil
}

// 生成动态的 patch 数据
func generateRequiredPodAffinityDynamicPatch(kind string, terms []interface{}) (map[string]interface{}, error) {
	// 打印rules
	klog.V(7).Infof("generateRequiredPodAffinityDynamicPatch rules:\n%+v len=%d\n", terms, len(terms))
	// [] len=0
	// 获取资源路径
	paths, err := getResourcePaths(kind)
	if err != nil {
		return nil, err
	}
	requiredAffinityPath := append(paths, "affinity", "podAffinity")

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

	// 单独处理删除时，terms为空的情况，要赋值为nil
	// 生成json 如下，那么需要将{}替换为null，为null才会删除
	// {
	//   "spec": {
	//     "jobTemplate": {
	//       "spec": {
	//         "template": {
	//           "spec": {
	//             "affinity": {
	//               "podAffinity": {}
	//             }
	//           }
	//         }
	//       }
	//     }
	//   }
	// }
	if len(terms) == 0 {
		current = nil
		return patch, nil
	}

	current["requiredDuringSchedulingIgnoredDuringExecution"] = terms

	return patch, nil
}
