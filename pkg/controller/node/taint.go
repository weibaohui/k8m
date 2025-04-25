package node

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
)

// ListTaint 获取某个节点上的污点
func ListTaint(c *gin.Context) {
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

// AddTaint 添加污点
func AddTaint(c *gin.Context) {
	if err := processTaint(c, "add"); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

// DeleteTaint 删除污点
func DeleteTaint(c *gin.Context) {
	if err := processTaint(c, "del"); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

// UpdateTaint 修改污点
func UpdateTaint(c *gin.Context) {
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
