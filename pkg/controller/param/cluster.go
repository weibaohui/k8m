package param

import (
	"fmt"
	"time"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/kom"
)

// @Summary 集群选项列表
// @Description 获取当前登录用户可选的集群列表（下拉选项）
// @Security BearerAuth
// @Success 200 {object} string
// @Router /params/cluster/option_list [get]
func (pc *Controller) ClusterOptionList(c *gin.Context) {
	user := amis.GetLoginOnlyUserName(c)

	clusters := service.ClusterService().AllClusters()

	if len(clusters) == 0 {
		amis.WriteJsonData(c, gin.H{
			"options": make([]map[string]string, 0),
		})
		return
	}
	if !service.UserService().IsUserPlatformAdmin(user) {
		userCluster, err := service.UserService().GetClusterNames(user)
		if err != nil {
			amis.WriteJsonData(c, gin.H{
				"options": make([]map[string]string, 0),
			})
			return
		}
		clusters = slice.Filter(clusters, func(index int, cluster *service.ClusterConfig) bool {
			return slice.Contain(userCluster, cluster.GetClusterID())
		})
	}

	var options []map[string]any
	for _, cluster := range clusters {
		name := cluster.GetClusterID()
		flag := "✅"
		if cluster.ClusterConnectStatus != constants.ClusterConnectStatusConnected {
			flag = "⚠️"
		}
		options = append(options, map[string]any{
			"label": fmt.Sprintf("%s %s", flag, name),
			"value": name,
			// "disabled": cluster.ServerVersion == "",
		})
	}

	amis.WriteJsonData(c, gin.H{
		"options": options,
	})
}

// @Summary 集群表格列表
// @Description 获取当前登录用户可见的集群详细信息（表格）
// @Security BearerAuth
// @Success 200 {object} string
// @Router /params/cluster/all [get]
func (pc *Controller) ClusterTableList(c *gin.Context) {
	user := amis.GetLoginOnlyUserName(c)

	clusters := service.ClusterService().AllClusters()
	if !service.UserService().IsUserPlatformAdmin(user) {
		userCluster, err := service.UserService().GetClusterNames(user)
		if err != nil {
			amis.WriteJsonData(c, gin.H{
				"options": make([]map[string]string, 0),
			})
			return
		}
		clusters = slice.Filter(clusters, func(index int, cluster *service.ClusterConfig) bool {
			return slice.Contain(userCluster, cluster.GetClusterID())
		})
	}
	// 增加cluster.NotAfter
	configs := service.ClusterService().ConnectedClusters() // 优化：移到循环外部
	for _, cluster := range clusters {
		// InCluster AWS
		if !(cluster.IsInCluster || cluster.IsAWSEKS) && slice.ContainBy(configs, func(item *service.ClusterConfig) bool {
			return item.ClusterID == cluster.ClusterID
		}) {
			cacheKey := fmt.Sprintf("%s/kubeconfig/not_after", cluster.ClusterID)
			if notAfter, err := utils.GetOrSetCache(kom.Cluster(cluster.ClusterID).ClusterCache(), cacheKey, 24*time.Hour, func() (time.Time, error) {
				return cluster.GetCertificateExpiry(), nil
			}); err == nil {
				cluster.NotAfter = &notAfter
			}
		}
		if cluster.NotAfter != nil && cluster.NotAfter.IsZero() {
			cluster.NotAfter = nil
		}
	}
	amis.WriteJsonData(c, clusters)
}
