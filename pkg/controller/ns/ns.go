package ns

import (
	"strings"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
)

func OptionList(c *gin.Context) {
	ctx := amis.GetContextWithUser(c)
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
	// 先看jwt登录用户中，是否有限制的ns
	_, _, clusterUserRoles := amis.GetLoginUserWithClusterRoles(c)
	if clusterUserRoles != nil {
		// 先筛选带有ns的授权列表
		clusterUserRoles = slice.Filter(clusterUserRoles, func(index int, item *models.ClusterUserRole) bool {
			return item.Namespaces != "" && item.Cluster == selectedCluster
		})
		if len(clusterUserRoles) > 0 {
			// 具有授权列表，摘取其中的ns
			for _, item := range clusterUserRoles {
				if item.Namespaces != "" {
					ns := strings.Split(item.Namespaces, ",")
					for _, n := range ns {
						list = append(list, map[string]string{
							"label": n,
							"value": n,
						})
					}
				}
			}
		} else {
			// 授权列表中，都没有指定ns，说明不需要按ns进行过滤
			// 那么读取集群中的ns
			var ns []v1.Namespace
			err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Namespace{}).List(&ns).Error
			if err != nil {
				amis.WriteJsonData(c, gin.H{
					"options": make([]map[string]string, 0),
				})
				return
			}

			for _, n := range ns {
				list = append(list, map[string]string{
					"label": n.Name,
					"value": n.Name,
				})
			}
		}

	}

	slice.SortBy(list, func(a, b map[string]string) bool {
		return a["label"] < b["label"]
	})
	amis.WriteJsonData(c, gin.H{
		"options": list,
	})
}
