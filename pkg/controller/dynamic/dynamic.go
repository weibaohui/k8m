package dynamic

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
	"sigs.k8s.io/yaml"
)

func List(c *gin.Context) {
	ns := c.Param("ns")
	group := c.Param("group")
	kind := c.Param("kind")
	version := c.Param("version")
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	nsList := []string{}
	if strings.Contains(ns, ",") {
		nsList = strings.Split(ns, ",")
	} else {
		nsList = []string{ns}
	}

	// 用于存储 JSON 数据的 map
	var jsonData map[string]interface{}
	if err := c.ShouldBindJSON(&jsonData); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var total int64
	var list []unstructured.Unstructured
	sql := kom.Cluster(selectedCluster).WithContext(ctx).
		RemoveManagedFields().
		Namespace(nsList...).
		GVK(group, version, kind)

	// 处理查询条件
	queryConditions := parseNestedJSON("", jsonData)
	queryConditions = slice.Filter(queryConditions, func(index int, item string) bool {
		return !strings.HasSuffix(item, "=")
	})

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
	err := sql.
		FillTotalCount(&total).
		List(&list).Error

	list = FillList(selectedCluster, kind, list)
	amis.WriteJsonListTotalWithError(c, total, list, err)
}

// FillList 定制填充list []unstructured.Unstructured列表
func FillList(selectedCluster string, kind string, list []unstructured.Unstructured) []unstructured.Unstructured {
	switch kind {
	case "Node":
		if service.ClusterService().GetNodeStatusAggregated(selectedCluster) {
			// 已缓存聚合状态，可以填充
			for i, _ := range list {
				item := list[i]
				item = service.NodeService().SetIPUsage(selectedCluster, item)
				item = service.NodeService().SetPodCount(selectedCluster, item)
				item = service.NodeService().SetAllocatedStatus(selectedCluster, item)
			}
		}
	case "Pod":
		if service.ClusterService().GetPodStatusAggregated(selectedCluster) {
			// 已缓存聚合状态，可以填充
			for i, _ := range list {
				item := list[i]
				item = service.PodService().SetAllocatedStatus(selectedCluster, item)
			}
		}
	}
	return list
}
func Event(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	kind := c.Param("kind")
	group := c.Param("group")
	version := c.Param("version")
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	apiVersion := fmt.Sprintf("%s", version)
	if group != "" {
		apiVersion = fmt.Sprintf("%s/%s", group, version)
	}

	fieldSelector := fmt.Sprintf("regarding.apiVersion=%s,regarding.kind=%s,regarding.name=%s,regarding.namespace=%s", apiVersion, kind, name, ns)

	var eventList []unstructured.Unstructured
	err := kom.Cluster(selectedCluster).
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
	selectedCluster := amis.GetSelectedCluster(c)

	var obj *unstructured.Unstructured

	err := kom.Cluster(selectedCluster).WithContext(ctx).RemoveManagedFields().Name(name).Namespace(ns).CRD(group, version, kind).Get(&obj).Error
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
	selectedCluster := amis.GetSelectedCluster(c)

	err := removeSingle(ctx, selectedCluster, kind, group, version, ns, name, false)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)

}
func removeSingle(ctx context.Context, selectedCluster, kind, group, version, ns, name string, force bool) error {
	if force {
		return kom.Cluster(selectedCluster).WithContext(ctx).Name(name).Namespace(ns).CRD(group, version, kind).ForceDelete().Error
	}
	return kom.Cluster(selectedCluster).WithContext(ctx).Name(name).Namespace(ns).CRD(group, version, kind).Delete().Error
}

// NamesPayload 定义结构体以匹配批量删除 JSON 结构
type NamesPayload struct {
	Names []string `json:"names"`
}

func BatchRemove(c *gin.Context) {
	kind := c.Param("kind")
	group := c.Param("group")
	version := c.Param("version")
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	var req struct {
		Names      []string `json:"name_list"`
		Namespaces []string `json:"ns_list"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	var err error
	for i := 0; i < len(req.Names); i++ {
		name := req.Names[i]
		ns := req.Namespaces[i]
		x := removeSingle(ctx, selectedCluster, kind, group, version, ns, name, false)
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
func BatchForceRemove(c *gin.Context) {
	kind := c.Param("kind")
	group := c.Param("group")
	version := c.Param("version")
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	var req struct {
		Names      []string `json:"name_list"`
		Namespaces []string `json:"ns_list"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	var err error
	for i := 0; i < len(req.Names); i++ {
		name := req.Names[i]
		ns := req.Namespaces[i]
		x := removeSingle(ctx, selectedCluster, kind, group, version, ns, name, true)
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

type yamlRequest struct {
	Yaml string `json:"yaml" binding:"required"`
}

func Save(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	kind := c.Param("kind")
	group := c.Param("group")
	version := c.Param("version")
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	var req yamlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	yamlStr := req.Yaml

	// 解析 Yaml 到 Unstructured 对象
	var obj unstructured.Unstructured
	if err := yaml.Unmarshal([]byte(yamlStr), &obj.Object); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	obj.SetName(name)
	obj.SetNamespace(ns)
	err := kom.Cluster(selectedCluster).WithContext(ctx).Name(name).Namespace(ns).CRD(group, version, kind).Update(&obj).Error
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
	selectedCluster := amis.GetSelectedCluster(c)

	var result []byte

	err := kom.Cluster(selectedCluster).WithContext(ctx).Name(name).Namespace(ns).CRD(group, version, kind).Describe(&result).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, string(result))
}

func UploadFile(c *gin.Context) {
	selectedCluster := amis.GetSelectedCluster(c)

	ctx := c.Request.Context()
	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		amis.WriteJsonError(c, fmt.Errorf("error retrieving file: %v", err))
		return
	}
	src, err := file.Open()
	if err != nil {
		amis.WriteJsonError(c, fmt.Errorf("error openning file: %v", err))
		return
	}
	defer src.Close()
	yamlBytes, err := io.ReadAll(src)
	yamlStr := string(yamlBytes)
	result := kom.Cluster(selectedCluster).WithContext(ctx).Applier().Apply(yamlStr)
	amis.WriteJsonOKMsg(c, strings.Join(result, "\n"))
}

func Apply(c *gin.Context) {
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	var req yamlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, fmt.Errorf("提取yaml错误。\n %v", err))
		return
	}
	yamlStr := req.Yaml
	result := kom.Cluster(selectedCluster).WithContext(ctx).Applier().Apply(yamlStr)
	amis.WriteJsonData(c, gin.H{
		"result": result,
	})

}
func Delete(c *gin.Context) {
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	var req yamlRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	yamlStr := req.Yaml
	result := kom.Cluster(selectedCluster).WithContext(ctx).Applier().Delete(yamlStr)
	amis.WriteJsonData(c, gin.H{
		"result": result,
	})
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

func UpdateLabels(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	kind := c.Param("kind")
	group := c.Param("group")
	version := c.Param("version")
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	var req struct {
		Labels map[string]string `json:"labels"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var obj *unstructured.Unstructured
	err := kom.Cluster(selectedCluster).WithContext(ctx).
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

func UpdateAnnotations(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	kind := c.Param("kind")
	group := c.Param("group")
	version := c.Param("version")
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	var req struct {
		Annotations map[string]interface{} `json:"annotations"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 部分key为k8m增加的指标数据，不是资源自身的注解，因此过滤掉。
	// last-applied-configuration是k8s管理的，不允许修改。
	var immutableKeys = []string{
		"cpu.request",
		"cpu.requestFraction",
		"cpu.limit",
		"cpu.limitFraction",
		"cpu.total",
		"memory.request",
		"memory.requestFraction",
		"memory.limit",
		"memory.limitFraction",
		"memory.total",
		"ip.usage.total",
		"ip.usage.used",
		"ip.usage.available",
		"pod.count.total",
		"pod.count.used",
		"pod.count.available",
		"kubectl.kubernetes.io/last-applied-configuration",
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
	err := kom.Cluster(selectedCluster).WithContext(ctx).
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

func GroupOptionList(c *gin.Context) {
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	var list []unstructured.Unstructured
	err := kom.Cluster(selectedCluster).WithContext(ctx).GVK(
		"apiextensions.k8s.io",
		"v1",
		"CustomResourceDefinition").
		WithCache(time.Second * 30).
		List(&list).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var groups []string
	for _, item := range list {
		group, found, err := unstructured.NestedString(item.Object, "spec", "group")
		if err != nil || !found {
			continue
		}

		groups = append(groups, group)
	}
	groups = slice.Unique(groups)
	var options []map[string]string
	for _, n := range groups {
		options = append(options, map[string]string{
			"label": n,
			"value": n,
		})
	}

	amis.WriteJsonData(c, gin.H{
		"options": options,
	})
}

func KindOptionList(c *gin.Context) {
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)
	g := c.Query("spec[group]")
	if g == "" {
		// 还没选group
		amis.WriteJsonData(c, gin.H{
			"options": make([]map[string]string, 0),
		})
		return
	}
	klog.V(2).Infof("spec[group]=%s", g)
	var list []unstructured.Unstructured
	err := kom.Cluster(selectedCluster).WithContext(ctx).GVK(
		"apiextensions.k8s.io",
		"v1",
		"CustomResourceDefinition").
		Where("`spec.group`=?", g).
		WithCache(time.Second * 30).
		List(&list).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var kinds []string
	for _, item := range list {
		kind, found, err := unstructured.NestedString(item.Object, "spec", "names", "kind")
		if err != nil || !found {
			continue
		}

		kinds = append(kinds, kind)
	}
	kinds = slice.Unique(kinds)
	var options []map[string]string
	for _, n := range kinds {
		options = append(options, map[string]string{
			"label": n,
			"value": n,
		})
	}

	amis.WriteJsonData(c, gin.H{
		"options": options,
	})
}
