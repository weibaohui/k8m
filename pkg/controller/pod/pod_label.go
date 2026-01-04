package pod

import (
	"github.com/duke-git/lancet/v2/slice"
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/response"
	"github.com/weibaohui/k8m/pkg/service"
)

type LabelController struct{}

func RegisterLabelRoutes(api chi.Router) {
	ctrl := &LabelController{}
	api.Get("/pod/labels/unique_labels", response.Adapter(ctrl.UniqueLabels))

}

// @Summary 获取Pod唯一标签键列表
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/pod/labels/unique_labels [get]
// UniqueLabels 返回当前集群中所有唯一的 Pod 标签键列表，格式化为前端可用的选项数组。
func (lc *LabelController) UniqueLabels(c *response.Context) {
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
	amis.WriteJsonData(c, response.H{
		"options": names,
	})
}
