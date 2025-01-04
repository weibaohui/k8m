package amis

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/service"
)

func GetSelectedCluster(c *gin.Context) string {
	selectedCluster, _ := c.Cookie("selectedCluster")
	if selectedCluster == "" {
		clusters := service.ClusterService().ConnectedClusters()
		if len(clusters) > 0 {
			c := clusters[0]
			selectedCluster = fmt.Sprintf("%s/%s", c.FileName, c.ContextName)
			return selectedCluster
		}
	}
	return selectedCluster
}
