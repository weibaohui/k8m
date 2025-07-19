package pod

import (
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
)

type ResourceController struct{}

func RegisterResourceRoutes(api *gin.RouterGroup) {
	ctrl := &ResourceController{}
	api.GET("/pod/usage/ns/:ns/name/:name", ctrl.Usage)
	api.GET("/pod/top/ns/:ns/list", ctrl.TopList)
}

func (rc *ResourceController) Usage(c *gin.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	usage, err := kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.Pod{}).
		Namespace(ns).
		Name(name).
		Ctl().Pod().ResourceUsageTable()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, usage)
}

// TopList 返回指定命名空间下所有 Pod 的资源使用情况（CPU、内存等），支持多命名空间查询，并以便于前端排序的格式输出。
func (rc *ResourceController) TopList(c *gin.Context) {
	ns := c.Param("ns")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	podMetrics, err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Pod{}).
		Namespace(strings.Split(ns, ",")...).
		WithCache(time.Second * 30).
		Ctl().Pod().Top()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 转换为map 前端排序使用，usage.cpu这种前端无法正确排序
	var result []map[string]string
	for _, item := range podMetrics {
		result = append(result, map[string]string{
			"name":            item.Name,
			"namespace":       item.Namespace,
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
