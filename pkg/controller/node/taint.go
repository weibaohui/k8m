package node

import (
	"fmt"
	"time"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
)

type TaintController struct{}

func RegisterTaintRoutes(api *gin.RouterGroup) {
	ctrl := &TaintController{}
	api.POST("/node/update_taints/name/:name", ctrl.Update)
	api.POST("/node/delete_taints/name/:name", ctrl.Delete)
	api.POST("/node/add_taints/name/:name", ctrl.Add)
	api.GET("/node/list_taints/name/:name", ctrl.ListByName)
	api.GET("/node/taints/list", ctrl.List)
}

// @Summary 获取所有节点上的污点
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/node/taints/list [get]
func (tc *TaintController) List(c *gin.Context) {
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
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

// @Summary 获取某个节点上的污点
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param name path string true "节点名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/node/list_taints/name/{name} [get]
func (tc *TaintController) ListByName(c *gin.Context) {
	name := c.Param("name")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var node v1.Node
	err = kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Node{}).Name(name).
		Get(&node).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	taints := node.Spec.Taints
	amis.WriteJsonList(c, taints)
}

type TaintInfo struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Effect string `json:"effect"`
}

// @Summary 添加污点
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param name path string true "节点名称"
// @Param body body TaintInfo true "污点信息"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/node/add_taints/name/{name} [post]
func (tc *TaintController) Add(c *gin.Context) {
	if err := processTaint(c, "add"); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

// @Summary 删除污点
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param name path string true "节点名称"
// @Param body body TaintInfo true "污点信息"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/node/delete_taints/name/{name} [post]
func (tc *TaintController) Delete(c *gin.Context) {
	if err := processTaint(c, "del"); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

// @Summary 修改污点
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param name path string true "节点名称"
// @Param body body TaintInfo true "污点信息"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/node/update_taints/name/{name} [post]
func (tc *TaintController) Update(c *gin.Context) {
	if err := processTaint(c, "modify"); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

func processTaint(c *gin.Context, mode string) error {
	name := c.Param("name")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		return err
	}

	var info TaintInfo
	err = c.ShouldBindJSON(&info)
	if err != nil {
		return err
	}

	var node v1.Node
	err = kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Node{}).Name(name).
		Get(&node).Error
	if err != nil {
		return err
	}
	// tanintFormat:="dedicated2=special-user:NoSchedule"
	// tanintFormat:="dedicated2:NoSchedule"
	tanintString := fmt.Sprintf("%s=%s:%s", info.Key, info.Value, info.Effect)
	if info.Value == "" {
		tanintString = fmt.Sprintf("%s:%s", info.Key, info.Effect)
	}
	switch mode {
	case "add", "modify":
		err = kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Node{}).Name(name).
			Ctl().Node().Taint(tanintString)
	case "del":
		err = kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Node{}).Name(name).
			Ctl().Node().UnTaint(tanintString)
	}

	return err
}
