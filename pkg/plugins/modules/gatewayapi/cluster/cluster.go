package cluster

import (
	"github.com/duke-git/lancet/v2/slice"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/response"
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/kom"
	"k8s.io/klog/v2"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func GatewayClassOptionList(c *response.Context) {
	ok, err := service.AuthService().EnsureUserIsLogined(c)
	if !ok {
		amis.WriteJsonError(c, err)
		return
	}

	klog.V(6).Infof("获取GatewayClass选项列表")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var list []gatewayapiv1.GatewayClass
	err = kom.Cluster(selectedCluster).WithContext(ctx).
		CRD("gateway.networking.k8s.io", "v1", "GatewayClass").
		Resource(&gatewayapiv1.GatewayClass{}).List(&list).Error
	if err != nil {
		amis.WriteJsonData(c, response.H{
			"options": make([]map[string]string, 0),
		})
		return
	}
	var names []map[string]string
	for _, n := range list {
		names = append(names, map[string]string{
			"label": n.Name,
			"value": n.Name,
		})
	}
	slice.SortBy(names, func(a, b map[string]string) bool {
		return a["label"] < b["label"]
	})
	amis.WriteJsonData(c, response.H{
		"options": names,
	})
}
