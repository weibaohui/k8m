package node

import (
	"time"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
)

func Drain(c *gin.Context) {
	name := c.Param("name")
	ctx := amis.GetContextWithUser(c)
	selectedCluster := amis.GetSelectedCluster(c)

	err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Node{}).Name(name).
		Ctl().Node().Drain()
	amis.WriteJsonErrorOrOK(c, err)
}
func Cordon(c *gin.Context) {
	name := c.Param("name")
	ctx := amis.GetContextWithUser(c)
	selectedCluster := amis.GetSelectedCluster(c)

	err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Node{}).Name(name).
		Ctl().Node().Cordon()
	amis.WriteJsonErrorOrOK(c, err)
}
func Usage(c *gin.Context) {
	name := c.Param("name")
	ctx := amis.GetContextWithUser(c)
	selectedCluster := amis.GetSelectedCluster(c)

	usage := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Node{}).Name(name).
		Ctl().Node().ResourceUsageTable()
	// todo 增加其他资源用量
	amis.WriteJsonData(c, usage)
}
func UnCordon(c *gin.Context) {
	name := c.Param("name")
	ctx := amis.GetContextWithUser(c)
	selectedCluster := amis.GetSelectedCluster(c)

	err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Node{}).Name(name).
		Ctl().Node().UnCordon()
	amis.WriteJsonErrorOrOK(c, err)
}

func BatchDrain(c *gin.Context) {
	ctx := amis.GetContextWithUser(c)
	selectedCluster := amis.GetSelectedCluster(c)

	var req struct {
		Names []string `json:"name_list"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var err error
	for i := 0; i < len(req.Names); i++ {
		name := req.Names[i]
		x := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Node{}).Name(name).
			Ctl().Node().Drain()
		if x != nil {
			klog.V(6).Infof("批量驱逐节点错误 %s %v", name, x)
			err = x
		}
	}

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

func BatchCordon(c *gin.Context) {
	ctx := amis.GetContextWithUser(c)
	selectedCluster := amis.GetSelectedCluster(c)

	var req struct {
		Names []string `json:"name_list"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var err error
	for i := 0; i < len(req.Names); i++ {
		name := req.Names[i]
		x := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Node{}).Name(name).
			Ctl().Node().Cordon()
		if x != nil {
			klog.V(6).Infof("批量隔离节点错误 %s %v", name, x)
			err = x
		}
	}

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

func BatchUnCordon(c *gin.Context) {
	ctx := amis.GetContextWithUser(c)
	selectedCluster := amis.GetSelectedCluster(c)

	var req struct {
		Names []string `json:"name_list"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var err error
	for i := 0; i < len(req.Names); i++ {
		name := req.Names[i]
		x := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Node{}).Name(name).
			Ctl().Node().UnCordon()
		if x != nil {
			klog.V(6).Infof("批量解除节点隔离错误 %s %v", name, x)
			err = x
		}
	}

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

func NameOptionList(c *gin.Context) {
	ctx := amis.GetContextWithUser(c)
	selectedCluster := amis.GetSelectedCluster(c)

	var list []unstructured.Unstructured
	err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Node{}).
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
func AllLabelList(c *gin.Context) {
	ctx := amis.GetContextWithUser(c)
	selectedCluster := amis.GetSelectedCluster(c)
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

// AllTaintList 获取所有节点上的污点
func AllTaintList(c *gin.Context) {
	ctx := amis.GetContextWithUser(c)
	selectedCluster := amis.GetSelectedCluster(c)

	var nodeList []*v1.Node
	err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Node{}).
		WithCache(time.Second * 30).
		List(&nodeList).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	var taintsList []*v1.Taint

	for _, node := range nodeList {
		for _, taint := range node.Spec.Taints {
			taintsList = append(taintsList, &taint)
		}
	}

	type table struct {
		Names  []string       `json:"names"`
		IPs    []string       `json:"ips"`
		Key    string         `json:"key"`
		Value  string         `json:"value"`
		Effect v1.TaintEffect `json:"effect"`
	}

	var resultList []*table

	for _, v := range taintsList {
		resultList = append(resultList, &table{
			Key:    v.Key,
			Value:  v.Value,
			Effect: v.Effect,
			Names:  make([]string, 0),
			IPs:    make([]string, 0),
		})
	}
	// 排重
	resultList = slice.UniqueByComparator(resultList, func(i, j *table) bool {
		return i.Key == j.Key && i.Value == j.Value && i.Effect == j.Effect
	})

	// 循环labelList，循环node，如果node的label和labelList的key相同，则将node name放入到labelList的node字段中
	for _, v := range resultList {
		for _, node := range nodeList {
			for _, taint := range node.Spec.Taints {
				if taint.Key == v.Key && taint.Value == v.Value && taint.Effect == v.Effect {
					v.Names = append(v.Names, node.Name)
					v.IPs = append(v.IPs, node.Status.Addresses[0].Address)
				}
			}
		}
	}

	amis.WriteJsonList(c, resultList)
}
