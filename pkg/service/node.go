package service

import (
	"fmt"
	"sync"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/utils"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
)

var nodeOnce sync.Once

type nodeService struct {
	Cache *ristretto.Cache[string, any]
}

func newNodeService() *nodeService {

	nodeOnce.Do(func() {
		klog.V(6).Infof("init localNodeService")
		cache, err := ristretto.NewCache(&ristretto.Config[string, any]{
			NumCounters: 1e7,     // number of keys to track frequency of (10M).
			MaxCost:     1 << 30, // maximum cost of cache (1GB).
			BufferItems: 64,      // number of keys per Get buffer.
		})
		if err != nil {
			klog.Errorf("Failed to create cache: %v", err)
		}
		localNodeService = &nodeService{Cache: cache}
	})

	return localNodeService
}

func (n *nodeService) SetIPUsage(item unstructured.Unstructured) unstructured.Unstructured {
	nodeName := item.GetName()
	total, used, available := kom.DefaultCluster().Name(nodeName).Ctl().Node().IPUsage()

	// 设置或追加 annotations
	utils.AddOrUpdateAnnotations(&item, map[string]string{
		"ip.usage.total":     fmt.Sprintf("%d", total),
		"ip.usage.used":      fmt.Sprintf("%d", used),
		"ip.usage.available": fmt.Sprintf("%d", available),
	})

	return item
}

// SetAllocatedStatus 设置节点的分配状态
func (n *nodeService) SetAllocatedStatus(item unstructured.Unstructured) unstructured.Unstructured {
	nodeName := item.GetName()
	table := kom.DefaultCluster().Name(nodeName).Ctl().Node().ResourceUsageTable()
	for _, row := range table {
		if row.ResourceType == "cpu" {
			// 设置或追加 annotations
			utils.AddOrUpdateAnnotations(&item, map[string]string{
				"cpu.request":         fmt.Sprintf("%s", row.Request),
				"cpu.requestFraction": fmt.Sprintf("%s", row.RequestFraction),
				"cpu.limit":           fmt.Sprintf("%s", row.Limit),
				"cpu.limitFraction":   fmt.Sprintf("%s", row.LimitFraction),
			})
		} else if row.ResourceType == "memory" {
			// 设置或追加 annotations
			utils.AddOrUpdateAnnotations(&item, map[string]string{
				"memory.request":         fmt.Sprintf("%s", row.Request),
				"memory.requestFraction": fmt.Sprintf("%s", row.RequestFraction),
				"memory.limit":           fmt.Sprintf("%s", row.Limit),
				"memory.limitFraction":   fmt.Sprintf("%s", row.LimitFraction),
			})
		}
	}

	return item
}
