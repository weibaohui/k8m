package dynamic

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	utils2 "github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

type ActionController struct{}

func RegisterActionRoutes(api *gin.RouterGroup) {
	ctrl := &ActionController{}
	api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name", ctrl.Fetch)                         // CRD
	api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name/json", ctrl.FetchJson)                // CRD
	api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name/event", ctrl.Event)                   // CRD
	api.GET("/:kind/group/:group/version/:version/ns/:ns/name/:name/hpa", ctrl.HPA)                       // CRD
	api.POST("/:kind/group/:group/version/:version/ns/:ns/name/:name/scale/replica/:replica", ctrl.Scale) // CRD
	api.POST("/:kind/group/:group/version/:version/remove/ns/:ns/name/:name", ctrl.Remove)                // CRD
	api.POST("/:kind/group/:group/version/:version/batch/remove", ctrl.BatchRemove)                       // CRD
	api.POST("/:kind/group/:group/version/:version/force_remove", ctrl.BatchForceRemove)                  // CRD
	api.POST("/:kind/group/:group/version/:version/update/ns/:ns/name/:name", ctrl.Save)                  // CRD       // CRD
	api.POST("/:kind/group/:group/version/:version/describe/ns/:ns/name/:name", ctrl.Describe)            // CRD
	api.POST("/:kind/group/:group/version/:version/list/ns/:ns", ctrl.List)                               // CRD
	api.POST("/:kind/group/:group/version/:version/list/ns/", ctrl.List)                                  // CRD
	api.POST("/:kind/group/:group/version/:version/list", ctrl.List)

}

// @Summary 获取资源列表
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param kind path string true "资源类型"
// @Param group path string true "资源组"
// @Param version path string true "资源版本"
// @Param ns path string true "命名空间"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/{kind}/group/{group}/version/{version}/list/ns/{ns} [post]
func (ac *ActionController) List(c *gin.Context) {
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

	var nsList []string
	if strings.Contains(ns, ",") {
		nsList = strings.Split(ns, ",")
	} else {
		nsList = []string{ns}
	}

	// 用于存储 JSON 数据的 map
	var jsonData map[string]interface{}
	if err = c.ShouldBindJSON(&jsonData); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var total int64
	var list []*unstructured.Unstructured
	sql := kom.Cluster(selectedCluster).WithContext(ctx).
		RemoveManagedFields().
		Namespace(nsList...).
		GVK(group, version, kind)

	if uniqueLabels, ok := jsonData["unique_labels"]; ok {
		if uniqueLabels != "" {
			delete(jsonData, "unique_labels")
			sql = sql.WithLabelSelector(uniqueLabels.(string))
		}
	}

	// 处理查询条件
	queryConditions := parseNestedJSON("", jsonData)
	queryConditions = slice.Filter(queryConditions, func(index int, item string) bool {
		return !strings.HasSuffix(item, "=")
	})

	// 检查jsonData中metadata.namespace是否在nsList中，如果不在，给出一个提示
	if metadata, ok := jsonData["metadata"]; ok {
		if metadataMap, ok := metadata.(map[string]interface{}); ok {
			if namespace, ok := metadataMap["namespace"]; ok {
				if namespaceStr, ok := namespace.(string); ok && namespaceStr != "" {
					// 检查namespaceStr是否在nsList中
					namespaceInList := slices.Contains(nsList, namespaceStr)
					if !namespaceInList {
						nsRangeError := fmt.Errorf("查询条件中的命名空间 '%s' 不在当前查询范围 [%v] 中，请重新选择", namespaceStr, strings.Join(nsList, ","))
						amis.WriteJsonError(c, nsRangeError)
						return
					}
				}
			}
		}
	}

	if len(queryConditions) > 0 {
		queryString := strings.Join(queryConditions, " and ")
		klog.V(6).Infof("sql string =%s", queryString)
		sql = sql.Where(queryString)
	}

	// 处理OrderBy,默认asc
	//  orderBy = 字段
	// orderDir = asc/desc/空
	orderBy, orderByOK := jsonData["orderBy"].(string)
	orderDir, orderDirOK := jsonData["orderDir"].(string)

	if orderByOK {
		if orderDirOK {
			sql = sql.Order(fmt.Sprintf("%s %s", orderBy, orderDir))
		} else {
			sql = sql.Order(fmt.Sprintf("%s asc", orderBy))
		}
	}
	// 取出 page 和 perPage
	// 取出 page 和 perPage
	page, pageOK := jsonData["page"].(float64) // JSON 数字会解析为 float64
	perPage, perPageOK := jsonData["perPage"].(float64)
	if pageOK {
		sql = sql.Limit(int(perPage))
	}
	if perPageOK {
		sql = sql.Offset((int(page) - 1) * int(perPage))
	}

	// 执行SQL
	err = sql.
		FillTotalCount(&total).
		List(&list).Error

	list = ac.fillList(selectedCluster, kind, list)
	amis.WriteJsonListTotalWithError(c, total, list, err)
}

// fillList 定制填充list []*unstructured.Unstructured列表
func (ac *ActionController) fillList(selectedCluster string, kind string, list []*unstructured.Unstructured) []*unstructured.Unstructured {
	switch kind {
	case "Node":
		if service.ClusterService().GetNodeStatusAggregated(selectedCluster) {
			// 已缓存聚合状态，可以填充
			for i := range list {
				item := list[i]
				service.NodeService().SetIPUsage(selectedCluster, item)
				service.NodeService().SetPodCount(selectedCluster, item)
				service.NodeService().SetAllocatedStatus(selectedCluster, item)
			}
		}
	case "Pod":
		if service.ClusterService().GetPodStatusAggregated(selectedCluster) {
			// 已缓存聚合状态，可以填充
			for i := range list {
				item := list[i]
				service.PodService().SetAllocatedStatusOnPod(selectedCluster, item)
			}
		}
	case "Namespace":
		if service.ClusterService().GetPodStatusAggregated(selectedCluster) {
			// 已缓存聚合状态，可以填充
			for i := range list {
				item := list[i]
				service.PodService().SetStatusCountOnNamespace(selectedCluster, item)
			}
		}
	case "StorageClass":
		if service.ClusterService().GetPVCStatusAggregated(selectedCluster) {
			// 已缓存聚合状态，可以填充
			for i := range list {
				item := list[i]
				service.StorageClassService().SetPVCCount(selectedCluster, item)
			}
		}
		if service.ClusterService().GetPVStatusAggregated(selectedCluster) {
			// 已缓存聚合状态，可以填充
			for i := range list {
				item := list[i]
				service.StorageClassService().SetPVCount(selectedCluster, item)
			}
		}
	case "IngressClass":
		if service.ClusterService().GetIngressStatusAggregated(selectedCluster) {
			// 已缓存聚合状态，可以填充
			for i := range list {
				item := list[i]
				service.IngressClassService().SetIngressCount(selectedCluster, item)
			}
		}
	}
	return list
}

// @Summary 获取资源事件
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param kind path string true "资源类型"
// @Param group path string true "资源组"
// @Param version path string true "资源版本"
// @Param ns path string true "命名空间"
// @Param name path string true "资源名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/{kind}/group/{group}/version/{version}/ns/{ns}/name/{name}/event [get]
func (ac *ActionController) Event(c *gin.Context) {
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

	apiVersion := version
	if group != "" {
		apiVersion = fmt.Sprintf("%s/%s", group, version)
	}

	fieldSelector := fmt.Sprintf("regarding.apiVersion=%s,regarding.kind=%s,regarding.name=%s,regarding.namespace=%s", apiVersion, kind, name, ns)

	var eventList []*unstructured.Unstructured
	err = kom.Cluster(selectedCluster).
		WithContext(ctx).
		RemoveManagedFields().
		Namespace(ns).
		GVK("events.k8s.io", "v1", "Event").
		List(&eventList, metav1.ListOptions{
			FieldSelector: fieldSelector,
		}).Error

	amis.WriteJsonListWithError(c, eventList, err)

}

// @Summary 获取资源YAML
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param kind path string true "资源类型"
// @Param group path string true "资源组"
// @Param version path string true "资源版本"
// @Param ns path string true "命名空间"
// @Param name path string true "资源名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/{kind}/group/{group}/version/{version}/ns/{ns}/name/{name} [get]
func (ac *ActionController) Fetch(c *gin.Context) {
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

	err = kom.Cluster(selectedCluster).WithContext(ctx).RemoveManagedFields().Name(name).Namespace(ns).CRD(group, version, kind).Get(&obj).Error
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

// @Summary 获取资源JSON
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param kind path string true "资源类型"
// @Param group path string true "资源组"
// @Param version path string true "资源版本"
// @Param ns path string true "命名空间"
// @Param name path string true "资源名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/{kind}/group/{group}/version/{version}/ns/{ns}/name/{name}/json [get]
func (ac *ActionController) FetchJson(c *gin.Context) {
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

	err = kom.Cluster(selectedCluster).WithContext(ctx).RemoveManagedFields().Name(name).Namespace(ns).CRD(group, version, kind).Get(&obj).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonData(c, obj)
}

// @Summary 删除单个资源
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param kind path string true "资源类型"
// @Param group path string true "资源组"
// @Param version path string true "资源版本"
// @Param ns path string true "命名空间"
// @Param name path string true "资源名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/{kind}/group/{group}/version/{version}/remove/ns/{ns}/name/{name} [post]
func (ac *ActionController) Remove(c *gin.Context) {
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

	err = ac.removeSingle(ctx, selectedCluster, kind, group, version, ns, name, false)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)

}
func (ac *ActionController) removeSingle(ctx context.Context, selectedCluster, kind, group, version, ns, name string, force bool) error {
	if force {
		return kom.Cluster(selectedCluster).WithContext(ctx).Name(name).Namespace(ns).CRD(group, version, kind).ForceDelete().Error
	}
	return kom.Cluster(selectedCluster).WithContext(ctx).Name(name).Namespace(ns).CRD(group, version, kind).Delete().Error
}

// NamesPayload 定义结构体以匹配批量删除 JSON 结构
type NamesPayload struct {
	Names []string `json:"names"`
}

// @Summary 批量删除资源
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param kind path string true "资源类型"
// @Param group path string true "资源组"
// @Param version path string true "资源版本"
// @Param name_list body []string true "资源名称列表"
// @Param ns_list body []string true "命名空间列表"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/{kind}/group/{group}/version/{version}/batch/remove [post]
func (ac *ActionController) BatchRemove(c *gin.Context) {
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
		Names      []string `json:"name_list"`
		Namespaces []string `json:"ns_list"`
	}
	if err = c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	for i := 0; i < len(req.Names); i++ {
		name := req.Names[i]
		ns := req.Namespaces[i]
		x := ac.removeSingle(ctx, selectedCluster, kind, group, version, ns, name, false)
		if x != nil {
			klog.V(6).Infof("batch remove %s error %s/%s %v", kind, ns, name, x)
			err = x
		}
	}

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonOK(c)
}

// @Summary 批量强制删除资源
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param kind path string true "资源类型"
// @Param group path string true "资源组"
// @Param version path string true "资源版本"
// @Param name_list body []string true "资源名称列表"
// @Param ns_list body []string true "命名空间列表"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/{kind}/group/{group}/version/{version}/force_remove [post]
func (ac *ActionController) BatchForceRemove(c *gin.Context) {
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
		Names      []string `json:"name_list"`
		Namespaces []string `json:"ns_list"`
	}
	if err = c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	for i := 0; i < len(req.Names); i++ {
		name := req.Names[i]
		ns := req.Namespaces[i]
		x := ac.removeSingle(ctx, selectedCluster, kind, group, version, ns, name, true)
		if x != nil {
			klog.V(6).Infof("batch force remove %s error %s/%s %v", kind, ns, name, x)
			err = x
		}
	}

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonOK(c)
}

// @Summary 更新资源
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param kind path string true "资源类型"
// @Param group path string true "资源组"
// @Param version path string true "资源版本"
// @Param ns path string true "命名空间"
// @Param name path string true "资源名称"
// @Param yaml body string true "资源YAML内容"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/{kind}/group/{group}/version/{version}/update/ns/{ns}/name/{name} [post]
func (ac *ActionController) Save(c *gin.Context) {
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

	var req yamlRequest
	if err = c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	yamlStr := req.Yaml

	// 解析 Yaml 到 Unstructured 对象

	obj := &unstructured.Unstructured{}
	var raw map[string]interface{}
	if err = yaml.Unmarshal([]byte(yamlStr), &raw); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	obj.Object = raw
	obj.SetName(name)
	obj.SetNamespace(ns)
	err = kom.Cluster(selectedCluster).WithContext(ctx).Name(name).Namespace(ns).CRD(group, version, kind).Update(&obj).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

// @Summary 描述资源
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param kind path string true "资源类型"
// @Param group path string true "资源组"
// @Param version path string true "资源版本"
// @Param ns path string true "命名空间"
// @Param name path string true "资源名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/{kind}/group/{group}/version/{version}/describe/ns/{ns}/name/{name} [post]
func (ac *ActionController) Describe(c *gin.Context) {
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

	var result []byte

	err = kom.Cluster(selectedCluster).WithContext(ctx).Name(name).Namespace(ns).CRD(group, version, kind).Describe(&result).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, string(result))
}

// @Summary 扩缩容资源
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param kind path string true "资源类型"
// @Param group path string true "资源组"
// @Param version path string true "资源版本"
// @Param ns path string true "命名空间"
// @Param name path string true "资源名称"
// @Param replica path string true "副本数"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/{kind}/group/{group}/version/{version}/ns/{ns}/name/{name}/scale/replica/{replica} [post]
func (ac *ActionController) Scale(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	replica := c.Param("replica")
	kind := c.Param("kind")
	group := c.Param("group")
	version := c.Param("version")
	r := utils2.ToInt32(replica)
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	err = kom.Cluster(selectedCluster).WithContext(ctx).
		CRD(group, version, kind).
		Namespace(ns).Name(name).
		Ctl().Scaler().Scale(r)
	amis.WriteJsonErrorOrOK(c, err)
}

// @Summary 获取资源HPA信息
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param kind path string true "资源类型"
// @Param group path string true "资源组"
// @Param version path string true "资源版本"
// @Param ns path string true "命名空间"
// @Param name path string true "资源名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/{kind}/group/{group}/version/{version}/ns/{ns}/name/{name}/hpa [get]
func (ac *ActionController) HPA(c *gin.Context) {
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
	hpa, err := kom.Cluster(selectedCluster).WithContext(ctx).
		CRD(group, version, kind).Namespace(ns).Name(name).
		Ctl().CRD().HPAList()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, hpa)
}

// 递归解析 JSON 数据
//
//	queryConditions := parseNestedJSON("", jsonData)
//
// // 示例 JSON 数据
//
//	jsonData := map[string]interface{}{
//		"page":    1,
//		"metadata": map[string]interface{}{
//			"name": "nginx",
//		},
//		"status": map[string]interface{}{
//			"phase": "Running",
//		},
//		"perPage": 10,
//	}
// 	queryString := strings.Join(queryConditions, "&")
// 输出: page=1&metadata.name=nginx&status.phase=Running&perPage=10

func parseNestedJSON(prefix string, data map[string]interface{}) []string {
	var result []string

	for key, value := range data {

		// 拼接当前路径
		currentKey := key
		if prefix != "" {
			currentKey = prefix + "." + key
		}
		// 分页参数跳过
		// ns name 已经单独在kom调用链中单独设定，不需要设置到where条件中
		ignoreKeys := []string{"page", "perPage", "pageDir", "orderDir", "orderBy", "keywords", "ns", "name"}
		if slice.Contain(ignoreKeys, currentKey) {
			continue
		}
		switch v := value.(type) {
		case map[string]interface{}:
			// 递归解析嵌套对象
			result = append(result, parseNestedJSON(currentKey, v)...)
		default:
			// 添加键值对
			if v == "" {
				// 没有值跳过
				continue
			}

			result = append(result, fmt.Sprintf("`%s` like '%%%v%%'", currentKey, v))
		}
	}

	return result
}
