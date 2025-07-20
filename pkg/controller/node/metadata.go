package node

import (
	"time"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type MetadataController struct{}

func RegisterMetadataRoutes(api *gin.RouterGroup) {
	ctrl := &MetadataController{}
	api.GET("/node/name/option_list", ctrl.NameOptionList)
	api.GET("/node/labels/list", ctrl.AllLabelList)
	api.GET("/node/labels/unique_labels", ctrl.UniqueLabels)
}
func (nc *MetadataController) NameOptionList(c *gin.Context) {
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var list []*unstructured.Unstructured
	err = kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Node{}).
		WithCache(time.Second * 30).
		List(&list).Error
	if err != nil {
		amis.WriteJsonData(c, gin.H{
			"options": make([]map[string]string, 0),
		})
		return
	}

	var names []string
	for _, n := range list {
		names = append(names, n.GetName())
	}
	slice.Sort(names, "asc")

	var options []map[string]string
	for _, n := range names {
		options = append(options, map[string]string{
			"label": n,
			"value": n,
		})
	}

	amis.WriteJsonData(c, gin.H{
		"options": options,
	})
}

// AllLabelList 获取所有节点上的标签
func (nc *MetadataController) AllLabelList(c *gin.Context) {
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	// 先拿到所有的lable列表
	// 通过lable的kv去匹配node，将node name放入到label 结构体中，方便选择时做出判断
	labels, err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Node{}).
		WithCache(time.Second * 30).Ctl().Node().AllNodeLabels()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	var nodeList []*v1.Node
	err = kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Node{}).
		WithCache(time.Second * 30).
		List(&nodeList).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	type table struct {
		Names []string `json:"names"`
		IPs   []string `json:"ips"`
		Key   string   `json:"key"`
		Value string   `json:"value"`
	}

	var labelList []*table

	for k, v := range labels {
		labelList = append(labelList, &table{
			Key:   k,
			Value: v,
			Names: make([]string, 0),
			IPs:   make([]string, 0),
		})
	}

	// 循环labelList，循环node，如果node的label和labelList的key相同，则将node name放入到labelList的node字段中
	for _, v := range labelList {
		for _, node := range nodeList {
			if node.Labels[v.Key] == v.Value {
				v.Names = append(v.Names, node.Name)
				v.IPs = append(v.IPs, node.Status.Addresses[0].Address)
			}
		}
	}

	amis.WriteJsonList(c, labelList)
}

// UniqueLabels 获取选定集群中所有唯一的节点标签键，并以选项列表形式返回。
func (nc *MetadataController) UniqueLabels(c *gin.Context) {
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	labels := service.NodeService().GetUniqueLabels(selectedCluster)

	var names []map[string]string
	for k := range labels {
		names = append(names, map[string]string{
			"label": k,
			"value": k,
		})
	}
	slice.SortBy(names, func(a, b map[string]string) bool {
		return a["label"] < b["label"]
	})
	amis.WriteJsonData(c, gin.H{
		"options": names,
	})
}
