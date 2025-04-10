package param

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/service"
)

// UserRole 获取当前用户的Role信息
func UserRole(c *gin.Context) {
	_, role := amis.GetLoginUser(c)
	clusters := amis.GetLoginUserClusters(c)
	var cluster string
	if len(clusters) == 1 {
		cluster = clusters[0]
	}
	if cluster == "" {
		if amis.IsCurrentUserPlatformAdmin(c) {
			cluster = service.ClusterService().FirstClusterID()
		}
	}

	amis.WriteJsonData(c, gin.H{
		"role":    role,
		"cluster": cluster,
	})
}
