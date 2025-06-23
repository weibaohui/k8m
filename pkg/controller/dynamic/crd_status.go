package dynamic

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
)

// CRDStatus 处理 HTTP 请求，返回当前选中集群是否支持 Gateway API 的状态。
func CRDStatus(c *gin.Context) {
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, gin.H{
		"IsGatewayAPISupported": kom.Cluster(selectedCluster).Status().IsGatewayAPISupported(),
		"IsOpenKruiseSupported": kom.Cluster(selectedCluster).Status().IsCRDSupportedByName("daemonsets.apps.kruise.io"),
		"IsIstioSupported":      kom.Cluster(selectedCluster).Status().IsCRDSupportedByName("sidecars.networking.istio.io"),
	})
}
