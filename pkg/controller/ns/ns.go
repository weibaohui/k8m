package ns

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
)

func OptionList(c *gin.Context) {
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	var ns []v1.Namespace
	err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Namespace{}).List(&ns).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	var list []map[string]string
	for _, n := range ns {
		list = append(list, map[string]string{
			"label": n.Name,
			"value": n.Name,
		})
	}
	list = append(list, map[string]string{
		"label": "全部",
		"value": "*",
	})
	amis.WriteJsonData(c, gin.H{
		"options": list,
	})
}
