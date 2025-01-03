package service

import (
	"fmt"
	"sync"
	"time"

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
	cacheKey := fmt.Sprintf("%s/%s", "IPUsage", nodeName)
	type usage struct {
		Total     int `json:"total"`
		Used      int `json:"used"`
		Available int `json:"available"`
	}
	u, _ := utils.GetOrSetCache(n.Cache, cacheKey, 10*time.Minute, func() (usage, error) {
		total, used, available := kom.DefaultCluster().Name(nodeName).Ctl().Node().IPUsage()

		return usage{
			Total:     total,
			Used:      used,
			Available: available,
		}, nil
	})
	// 设置或追加 annotations
	utils.AddOrUpdateAnnotations(&item, map[string]string{
		"ip.usage.total":     fmt.Sprintf("%d", u.Total),
		"ip.usage.used":      fmt.Sprintf("%d", u.Used),
		"ip.usage.available": fmt.Sprintf("%d", u.Available),
	})

	return item
}

// SetAllocatedStatus 设置节点的分配状态
func (n *nodeService) SetAllocatedStatus(item unstructured.Unstructured) unstructured.Unstructured {
	nodeName := item.GetName()
	cacheKey := fmt.Sprintf("%s/%s", "AllocatedStatus", nodeName)
	table, _ := utils.GetOrSetCache(n.Cache, cacheKey, 10*time.Minute, func() ([]*kom.ResourceUsageRow, error) {
		tb := kom.DefaultCluster().Name(nodeName).Ctl().Node().ResourceUsageTable()
		return tb, nil
	})

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
