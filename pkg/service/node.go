package service

import (
	"fmt"
	"time"

	utils2 "github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/flag"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
)

type ipUsage struct {
	Total     int `json:"total"`
	Used      int `json:"used"`
	Available int `json:"available"`
}

func (n *nodeService) getTTL() time.Duration {
	cfg := flag.Init()
	if cfg.ResourceCacheTimeout > 0 {
		return time.Duration(cfg.ResourceCacheTimeout) * time.Second
	}
	return 1 * time.Minute
}

func (n *nodeService) SetIPUsage(selectedCluster string, item unstructured.Unstructured) unstructured.Unstructured {
	nodeName := item.GetName()
	u, err := n.CacheIPUsage(selectedCluster, nodeName)
	if err != nil {
		return item
	}
	// 设置或追加 annotations
	utils.AddOrUpdateAnnotations(&item, map[string]string{
		"ip.usage.total":     fmt.Sprintf("%d", u.Total),
		"ip.usage.used":      fmt.Sprintf("%d", u.Used),
		"ip.usage.available": fmt.Sprintf("%d", u.Available),
	})

	return item
}
func (n *nodeService) SetPodCount(selectedCluster string, item unstructured.Unstructured) unstructured.Unstructured {
	nodeName := item.GetName()
	u, err := n.CachePodCount(selectedCluster, nodeName)
	if err != nil {
		return item
	}
	// 设置或追加 annotations
	utils.AddOrUpdateAnnotations(&item, map[string]string{
		"pod.count.total":     fmt.Sprintf("%d", u.Total),
		"pod.count.used":      fmt.Sprintf("%d", u.Used),
		"pod.count.available": fmt.Sprintf("%d", u.Available),
	})

	return item
}

// SetAllocatedStatus 设置节点的分配状态
func (n *nodeService) SetAllocatedStatus(selectedCluster string, item unstructured.Unstructured) unstructured.Unstructured {
	name := item.GetName()
	table, err := n.CacheAllocatedStatus(selectedCluster, name)
	if err != nil {
		return item
	}
	for _, row := range table {
		if row.ResourceType == "cpu" {
			// 设置或追加 annotations
			utils.AddOrUpdateAnnotations(&item, map[string]string{
				"cpu.request":          fmt.Sprintf("%s", row.Request),
				"cpu.requestFraction":  fmt.Sprintf("%s", row.RequestFraction),
				"cpu.limit":            fmt.Sprintf("%s", row.Limit),
				"cpu.limitFraction":    fmt.Sprintf("%s", row.LimitFraction),
				"cpu.realtime":         fmt.Sprintf("%s", row.Realtime),
				"cpu.realtimeFraction": fmt.Sprintf("%s", row.RealtimeFraction),
			})
		} else if row.ResourceType == "memory" {
			// 设置或追加 annotations
			utils.AddOrUpdateAnnotations(&item, map[string]string{
				"memory.request":          fmt.Sprintf("%s", row.Request),
				"memory.requestFraction":  fmt.Sprintf("%s", row.RequestFraction),
				"memory.limit":            fmt.Sprintf("%s", row.Limit),
				"memory.limitFraction":    fmt.Sprintf("%s", row.LimitFraction),
				"memory.realtime":         fmt.Sprintf("%s", row.Realtime),
				"memory.realtimeFraction": fmt.Sprintf("%s", row.RealtimeFraction),
			})
		}
	}

	return item
}
func (n *nodeService) SyncNodeStatus(selectedCluster string) {
	klog.V(6).Infof("Sync Node Status")
	ctx := utils2.GetContextWithAdmin()
	var nodes []v1.Node
	err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Node{}).WithCache(n.getTTL()).List(&nodes).Error
	if err != nil {
		klog.Errorf("监听Node失败:%v", err)
	}
	for _, node := range nodes {
		_, _ = n.CacheIPUsage(selectedCluster, node.Name)
		_, _ = n.CachePodCount(selectedCluster, node.Name)
		_, _ = n.CacheAllocatedStatus(selectedCluster, node.Name)
	}
	ClusterService().SetNodeStatusAggregated(selectedCluster, true)
}

func (n *nodeService) CacheIPUsage(selectedCluster string, nodeName string) (ipUsage, error) {
	cacheKey := fmt.Sprintf("%s/%s", "NodeIPUsage", nodeName)
	ctx := utils2.GetContextWithAdmin()
	return utils.GetOrSetCache(kom.Cluster(selectedCluster).ClusterCache(), cacheKey, n.getTTL(), func() (ipUsage, error) {
		total, used, available := kom.Cluster(selectedCluster).WithContext(ctx).Name(nodeName).WithCache(n.getTTL()).Ctl().Node().IPUsage()
		return ipUsage{
			Total:     total,
			Used:      used,
			Available: available,
		}, nil
	})
}
func (n *nodeService) CachePodCount(selectedCluster string, nodeName string) (ipUsage, error) {
	cacheKey := fmt.Sprintf("%s/%s", "NodePodCount", nodeName)
	ctx := utils2.GetContextWithAdmin()
	return utils.GetOrSetCache(kom.Cluster(selectedCluster).ClusterCache(), cacheKey, n.getTTL(), func() (ipUsage, error) {
		total, used, available := kom.Cluster(selectedCluster).WithContext(ctx).Name(nodeName).WithCache(n.getTTL()).Ctl().Node().PodCount()
		return ipUsage{
			Total:     total,
			Used:      used,
			Available: available,
		}, nil
	})
}

func (n *nodeService) CacheAllocatedStatus(selectedCluster string, nodeName string) ([]*kom.ResourceUsageRow, error) {
	cacheKey := fmt.Sprintf("%s/%s", "NodeAllocatedStatus", nodeName)
	ctx := utils2.GetContextWithAdmin()
	return utils.GetOrSetCache(kom.Cluster(selectedCluster).ClusterCache(), cacheKey, n.getTTL(), func() ([]*kom.ResourceUsageRow, error) {
		tb, err := kom.Cluster(selectedCluster).WithContext(ctx).Name(nodeName).WithCache(n.getTTL()).Resource(&v1.Node{}).Ctl().Node().ResourceUsageTable()
		return tb, err
	})
}
func (n *nodeService) RemoveNodeStatusCache(selectedCluster string, nodeName string) {
	NodeAllocatedStatusKey := fmt.Sprintf("%s/%s", "NodeAllocatedStatus", nodeName)
	NodeIPUsageKey := fmt.Sprintf("%s/%s", "NodeIPUsage", nodeName)
	NodePodCountKey := fmt.Sprintf("%s/%s", "NodePodCount", nodeName)
	keys := []string{NodeAllocatedStatusKey, NodeIPUsageKey, NodePodCountKey}
	for _, k := range keys {
		kom.Cluster(selectedCluster).ClusterCache().Del(k)
	}
}
