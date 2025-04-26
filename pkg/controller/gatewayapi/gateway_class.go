package gatewayapi

import (
	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
	gatewayapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func GatewayClassOptionList(c *gin.Context) {
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
		amis.WriteJsonData(c, gin.H{
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
	amis.WriteJsonData(c, gin.H{
		"options": names,
	})
}
