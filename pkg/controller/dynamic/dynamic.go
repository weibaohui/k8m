package dynamic

import (
	"context"
	"fmt"
	"io"
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

func List(c *gin.Context) {
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

	list = FillList(selectedCluster, kind, list)
	amis.WriteJsonListTotalWithError(c, total, list, err)
}

// FillList 定制填充list []unstructured.Unstructured列表
func FillList(selectedCluster string, kind string, list []unstructured.Unstructured) []unstructured.Unstructured {
	switch kind {
	case "Node":
		if service.ClusterService().GetNodeStatusAggregated(selectedCluster) {
			// 已缓存聚合状态，可以填充
			for i := range list {
				item := list[i]
				item = service.NodeService().SetIPUsage(selectedCluster, item)
				item = service.NodeService().SetPodCount(selectedCluster, item)
				item = service.NodeService().SetAllocatedStatus(selectedCluster, item)
			}
		}
	case "Pod":
		if service.ClusterService().GetPodStatusAggregated(selectedCluster) {
			// 已缓存聚合状态，可以填充
			for i := range list {
				item := list[i]
				item = service.PodService().SetAllocatedStatusOnPod(selectedCluster, item)
			}
		}
	case "Namespace":
		if service.ClusterService().GetPodStatusAggregated(selectedCluster) {
			// 已缓存聚合状态，可以填充
			for i := range list {
				item := list[i]
				item = service.PodService().SetStatusCountOnNamespace(selectedCluster, item)
			}
		}
	case "StorageClass":
		if service.ClusterService().GetPVCStatusAggregated(selectedCluster) {
			// 已缓存聚合状态，可以填充
			for i := range list {
				item := list[i]
				item = service.StorageClassService().SetPVCCount(selectedCluster, item)
			}
		}
		if service.ClusterService().GetPVStatusAggregated(selectedCluster) {
			// 已缓存聚合状态，可以填充
			for i := range list {
				item := list[i]
				item = service.StorageClassService().SetPVCount(selectedCluster, item)
			}
		}
	case "IngressClass":
		if service.ClusterService().GetIngressStatusAggregated(selectedCluster) {
			// 已缓存聚合状态，可以填充
			for i := range list {
				item := list[i]
				item = service.IngressClassService().SetIngressCount(selectedCluster, item)
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
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	apiVersion := fmt.Sprintf("%s", version)
	if group != "" {
		apiVersion = fmt.Sprintf("%s/%s", group, version)
	}

	fieldSelector := fmt.Sprintf("regarding.apiVersion=%s,regarding.kind=%s,regarding.name=%s,regarding.namespace=%s", apiVersion, kind, name, ns)

	var eventList []unstructured.Unstructured
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
func Fetch(c *gin.Context) {
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
func FetchJson(c *gin.Context) {
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
func Remove(c *gin.Context) {
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

	err = removeSingle(ctx, selectedCluster, kind, group, version, ns, name, false)
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
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
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
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
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
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

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
	err = kom.Cluster(selectedCluster).WithContext(ctx).Name(name).Namespace(ns).CRD(group, version, kind).Update(&obj).Error
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

func UploadFile(c *gin.Context) {
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	ctx := amis.GetContextWithUser(c)
	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		amis.WriteJsonError(c, fmt.Errorf("获取上传的文件错误。\n %v", err))
		return
	}
	src, err := file.Open()
	if err != nil {
		amis.WriteJsonError(c, fmt.Errorf("打开上传的文件错误。\n %v", err))
		return
	}
	defer src.Close()
	yamlBytes, err := io.ReadAll(src)
	if err != nil {
		amis.WriteJsonError(c, fmt.Errorf("读取上传的文件内容错误。\n %v", err))
		return
	}
	yamlStr := string(yamlBytes)
	result := kom.Cluster(selectedCluster).WithContext(ctx).Applier().Apply(yamlStr)
	amis.WriteJsonOKMsg(c, strings.Join(result, "\n"))
}

func Apply(c *gin.Context) {
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

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
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

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

func Scale(c *gin.Context) {
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
func HPA(c *gin.Context) {
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
