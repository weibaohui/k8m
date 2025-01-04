package service

import (
	"fmt"
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/robfig/cron/v3"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
)

// nodeStatusTTL 节点状态缓存时间
// 要跟watch中的定时处理器保持一致
var nodeStatusTTL = 5 * time.Minute

type nodeService struct {
}
type ipUsage struct {
	Total     int `json:"total"`
	Used      int `json:"used"`
	Available int `json:"available"`
}

func (n *nodeService) SetIPUsage(cache *ristretto.Cache[string, any], item unstructured.Unstructured) unstructured.Unstructured {
	nodeName := item.GetName()
	u, err := n.CacheIPUsage(cache, nodeName)
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
func (n *nodeService) SetPodCount(cache *ristretto.Cache[string, any], item unstructured.Unstructured) unstructured.Unstructured {
	nodeName := item.GetName()
	u, err := n.CachePodCount(cache, nodeName)
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
func (n *nodeService) SetAllocatedStatus(cache *ristretto.Cache[string, any], item unstructured.Unstructured) unstructured.Unstructured {
	name := item.GetName()
	table, err := n.CacheAllocatedStatus(cache, name)
	if err != nil {
		return item
	}
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
func (n *nodeService) SyncNodeStatus() {
	klog.V(6).Infof("Sync Node Status")
	var nodes []v1.Node
	cache := kom.DefaultCluster().ClusterCache()
	err := kom.DefaultCluster().Resource(&v1.Node{}).WithCache(nodeStatusTTL).List(&nodes)
	if err != nil {
		klog.Errorf("Error watch node:%v", err)
	}
	for _, node := range nodes {
		_, _ = n.CacheIPUsage(cache, node.Name)
		_, _ = n.CachePodCount(cache, node.Name)
		_, _ = n.CacheAllocatedStatus(cache, node.Name)
	}
}
func (n *nodeService) Watch() error {
	go func() {
		// 先执行一次
		n.SyncNodeStatus()
	}()
	// 设置一个定时器，后台不断更新node状态
	_, err := cron.New().AddFunc("@every 5m", func() {
		n.SyncNodeStatus()
	})

	if err != nil {
		return err
	}
	klog.V(6).Infof("新增节点状态定时更新任务【@every 5m】")
	return nil
}
func (n *nodeService) CacheIPUsage(cache *ristretto.Cache[string, any], nodeName string) (ipUsage, error) {
	cacheKey := fmt.Sprintf("%s/%s", "NodeIPUsage", nodeName)
	return utils.GetOrSetCache(cache, cacheKey, nodeStatusTTL, func() (ipUsage, error) {
		total, used, available := kom.DefaultCluster().Name(nodeName).WithCache(nodeStatusTTL).Ctl().Node().IPUsage()
		return ipUsage{
			Total:     total,
			Used:      used,
			Available: available,
		}, nil
	})
}
func (n *nodeService) CachePodCount(cache *ristretto.Cache[string, any], nodeName string) (ipUsage, error) {
	cacheKey := fmt.Sprintf("%s/%s", "NodePodCount", nodeName)
	return utils.GetOrSetCache(cache, cacheKey, nodeStatusTTL, func() (ipUsage, error) {
		total, used, available := kom.DefaultCluster().Name(nodeName).WithCache(nodeStatusTTL).Ctl().Node().PodCount()
		return ipUsage{
			Total:     total,
			Used:      used,
			Available: available,
		}, nil
	})
}

func (n *nodeService) CacheAllocatedStatus(cache *ristretto.Cache[string, any], nodeName string) ([]*kom.ResourceUsageRow, error) {
	cacheKey := fmt.Sprintf("%s/%s", "NodeAllocatedStatus", nodeName)
	return utils.GetOrSetCache(cache, cacheKey, nodeStatusTTL, func() ([]*kom.ResourceUsageRow, error) {
		tb := kom.DefaultCluster().Name(nodeName).WithCache(nodeStatusTTL).Resource(&v1.Node{}).Ctl().Node().ResourceUsageTable()
		return tb, nil
	})
}
func (n *nodeService) RemoveNodeStatusCache(cache *ristretto.Cache[string, any], nodeName string) {
	NodeAllocatedStatusKey := fmt.Sprintf("%s/%s", "NodeAllocatedStatus", nodeName)
	NodeIPUsageKey := fmt.Sprintf("%s/%s", "NodeIPUsage", nodeName)
	NodePodCountKey := fmt.Sprintf("%s/%s", "NodePodCount", nodeName)
	keys := []string{NodeAllocatedStatusKey, NodeIPUsageKey, NodePodCountKey}
	for _, k := range keys {
		cache.Del(k)
	}
}
