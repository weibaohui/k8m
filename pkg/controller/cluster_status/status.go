package cluster_status

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
)

func ClusterResourceCount(c *gin.Context) {
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
