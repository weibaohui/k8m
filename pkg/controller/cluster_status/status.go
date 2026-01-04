package cluster_status

import (
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/response"
	"github.com/weibaohui/kom/kom"
)

type ClusterController struct{}

// 从 gin 切换到 chi，使用 chi.Router 替代 gin.RouterGroup
func RegisterClusterRoutes(r chi.Router) {
	ctrl := &ClusterController{}
	r.Get("/status/resource_count/cache_seconds/{cache}", response.Adapter(ctrl.ClusterResourceCount))
}

// @Summary 获取集群资源数量统计
// @Security BearerAuth
// @Param cluster path string true "集群名称"
// @Param cache path string true "缓存时间（秒）"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/status/resource_count/cache_seconds/{cache} [get]
func (cc *ClusterController) ClusterResourceCount(c *response.Context) {
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	cacheStr := c.Param("cache")
	if cacheStr == "" {
		cacheStr = "30"
	}
	cache := utils.ToInt(cacheStr)
	sm, err := kom.Cluster(selectedCluster).Status().GetResourceCountSummary(cache)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 将sm的kv进行转换
	var result = make([]struct {
		Count    int
		Group    string
		Version  string
		Resource string
	}, 0)
	for gvr, count := range sm {
		result = append(result,
			struct {
				Count    int
				Group    string
				Version  string
				Resource string
			}{Count: count,
				Group:    gvr.Group,
				Version:  gvr.Version,
				Resource: gvr.Resource})
	}
	amis.WriteJsonData(c, result)
}
