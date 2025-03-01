package helm

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/helm"
	"github.com/weibaohui/k8m/pkg/service"
)

func ListReleaseHistory(c *gin.Context) {
	// ctx := amis.GetContextWithUser(c)
	selectedCluster := amis.GetSelectedCluster(c)
	restConfig := service.ClusterService().GetClusterByID(selectedCluster).GetRestConfig()
	h, err := helm.New(restConfig)
	if err != nil {
		amis.WriteJsonError(c, err)
	}
	history, err := h.GetReleaseHistory("haproxy-r")
	if err != nil {
		amis.WriteJsonError(c, err)
	}
	amis.WriteJsonData(c, history)
}
