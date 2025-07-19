package pod

import (
	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/service"
)

type LabelController struct{}

func RegisterLabelRoutes(api *gin.RouterGroup) {
	ctrl := &LabelController{}
	api.GET("/pod/labels/unique_labels", ctrl.UniqueLabels)

}

// UniqueLabels 返回当前集群中所有唯一的 Pod 标签键列表，格式化为前端可用的选项数组。
func (lc *LabelController) UniqueLabels(c *gin.Context) {
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	labels := service.PodService().GetUniquePodLabels(selectedCluster)

	var names []map[string]string
	for k := range labels {
		names = append(names, map[string]string{
			"label": k,
			"value": k,
		})
	}
	slice.SortBy(names, func(a, b map[string]string) bool {
		return a["label"] < b["label"]
	})
	amis.WriteJsonData(c, gin.H{
		"options": names,
	})
}
