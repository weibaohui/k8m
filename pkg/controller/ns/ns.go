package ns

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
)

func OptionList(c *gin.Context) {
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	var list []map[string]string
	// 先判断kubeconfig中是否限制了namespace
	// 1、如果限制了，那么从cluster 实例中取
	// 2、如果没有限制，那么从集群中取
	cluster := service.ClusterService().GetClusterByID(selectedCluster)
	if cluster != nil && cluster.Namespace != "" {
		list = append(list, map[string]string{
			"label": cluster.Namespace,
			"value": cluster.Namespace,
		})
		amis.WriteJsonData(c, gin.H{
			"options": list,
		})
		return
	}

	// 没有指定的情况
	var ns []v1.Namespace
	err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Namespace{}).List(&ns).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	list = append(list, map[string]string{
		"label": "全部",
		"value": "*",
	})
	for _, n := range ns {
		list = append(list, map[string]string{
			"label": n.Name,
			"value": n.Name,
		})
	}

	amis.WriteJsonData(c, gin.H{
		"options": list,
	})
}
