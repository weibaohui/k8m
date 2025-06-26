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

func ClusterOptionList(c *gin.Context) {
	user, _ := amis.GetLoginUser(c)

	clusters := service.ClusterService().AllClusters()

	if len(clusters) == 0 {
		amis.WriteJsonData(c, gin.H{
			"options": make([]map[string]string, 0),
		})
		return
	}
	if !amis.IsCurrentUserPlatformAdmin(c) {
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

	var options []map[string]interface{}
	for _, cluster := range clusters {
		name := cluster.GetClusterID()
		flag := "✅"
		if cluster.ClusterConnectStatus != constants.ClusterConnectStatusConnected {
			flag = "⚠️"
		}
		options = append(options, map[string]interface{}{
			"label": fmt.Sprintf("%s %s", flag, name),
			"value": name,
			// "disabled": cluster.ServerVersion == "",
		})
	}

	amis.WriteJsonData(c, gin.H{
		"options": options,
	})
}

func ClusterTableList(c *gin.Context) {
	user, _ := amis.GetLoginUser(c)

	clusters := service.ClusterService().AllClusters()
	if !amis.IsCurrentUserPlatformAdmin(c) {
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
	for _, cluster := range clusters {
		cacheKey := fmt.Sprintf("%s/kubeconfig/not_after", cluster.ClusterID)
		notAfter, err := utils.GetOrSetCache(kom.Cluster(cluster.ClusterID).ClusterCache(), cacheKey, 24*time.Hour, func() (time.Time, error) {
			return cluster.GetNotAfter(), nil
		})
		if err == nil {
			cluster.NotAfter = notAfter
		}
	}
	amis.WriteJsonData(c, clusters)
}
