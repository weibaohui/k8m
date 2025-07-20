package node

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
)

type ResourceController struct{}

func RegisterResourceRoutes(api *gin.RouterGroup) {
	ctrl := &ResourceController{}
	api.GET("/node/top/list", ctrl.TopList)
	api.GET("/node/usage/name/:name", ctrl.Usage)
}
func (nc *ResourceController) Usage(c *gin.Context) {
	name := c.Param("name")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	usage, err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Node{}).Name(name).
		Ctl().Node().ResourceUsageTable()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	// todo 增加其他资源用量
	amis.WriteJsonData(c, usage)
}

// TopList 返回所有节点的资源使用率（top指标），包括CPU和内存的用量及其数值化表示，便于前端排序和展示。
func (nc *ResourceController) TopList(c *gin.Context) {
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	nodeMetrics, err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Node{}).
		WithCache(time.Second * 30).
		Ctl().Node().Top()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	// 转换为map 前端排序使用，usage.cpu这种前端无法正确排序
	var result []map[string]string
	for _, item := range nodeMetrics {
		result = append(result, map[string]string{
			"name":            item.Name,
			"cpu":             item.Usage.CPU,
			"memory":          item.Usage.Memory,
			"cpu_nano":        fmt.Sprintf("%d", item.Usage.CPUNano),
			"memory_byte":     fmt.Sprintf("%d", item.Usage.MemoryByte),
			"cpu_fraction":    item.Usage.CPUFraction,
			"memory_fraction": item.Usage.MemoryFraction,
		})
	}
	amis.WriteJsonList(c, result)
}
