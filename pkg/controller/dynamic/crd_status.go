package dynamic

import (
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/response"
	"github.com/weibaohui/kom/kom"
)

// @Summary 获取CRD状态信息
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/crd/status [get]
// CRDStatus 处理 HTTP 请求，返回当前选中集群是否支持 Gateway API 的状态。
func (cc *CRDController) CRDStatus(c *response.Context) {
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, response.H{
		"IsGatewayAPISupported": kom.Cluster(selectedCluster).Status().IsGatewayAPISupported(),
		"IsOpenKruiseSupported": kom.Cluster(selectedCluster).Status().IsCRDSupportedByName("daemonsets.apps.kruise.io"),
		"IsIstioSupported":      kom.Cluster(selectedCluster).Status().IsCRDSupportedByName("sidecars.networking.istio.io"),
	})
}
