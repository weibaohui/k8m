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
	cacheKey := fmt.Sprintf("%s/%s", "NodeIPUsage", nodeName)
	u, _ := utils.GetOrSetCache(cache, cacheKey, nodeStatusTTL, func() (ipUsage, error) {
		total, used, available := kom.DefaultCluster().Name(nodeName).WithCache(nodeStatusTTL).Ctl().Node().IPUsage()
		return ipUsage{
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
func (n *nodeService) SetAllocatedStatus(cache *ristretto.Cache[string, any], item unstructured.Unstructured) unstructured.Unstructured {
	// todo改为后台周期性获取统计数据
	// todo 按集群进行处理，从kom里面获取cache使用。而不是自建，因为kom的cache是绑定集群的。
	name := item.GetName()
	version := item.GetResourceVersion()
	ns := item.GetNamespace()
	cacheKey := fmt.Sprintf("%s/%s/%s/%s", "NodeAllocatedStatus", ns, name, version)
	table, _ := utils.GetOrSetCache(cache, cacheKey, nodeStatusTTL, func() ([]*kom.ResourceUsageRow, error) {
		tb := kom.DefaultCluster().Name(name).WithCache(nodeStatusTTL).Ctl().Node().ResourceUsageTable()
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
func (n *nodeService) SyncNodeStatus() {
	klog.V(6).Infof("Sync Node Status")
	var nodes []v1.Node
	cache := kom.DefaultCluster().ClusterCache()
	err := kom.DefaultCluster().Resource(&v1.Node{}).List(&nodes)
	if err != nil {
		klog.Errorf("Error watch node:%v", err)
	}
	for _, node := range nodes {
		n.CacheIPUsage(cache, &node)
		n.CacheAllocatedStatus(cache, &node)
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
func (n *nodeService) CacheIPUsage(cache *ristretto.Cache[string, any], item *v1.Node) {
	nodeName := item.GetName()
	cacheKey := fmt.Sprintf("%s/%s", "NodeIPUsage", nodeName)
	_, _ = utils.GetOrSetCache(cache, cacheKey, nodeStatusTTL, func() (ipUsage, error) {
		total, used, available := kom.DefaultCluster().Name(nodeName).WithCache(nodeStatusTTL).Ctl().Node().IPUsage()
		return ipUsage{
			Total:     total,
			Used:      used,
			Available: available,
		}, nil
	})
}
func (n *nodeService) RemoveCacheIPUsage(cache *ristretto.Cache[string, any], item *v1.Node) {
	nodeName := item.GetName()
	cacheKey := fmt.Sprintf("%s/%s", "NodeIPUsage", nodeName)
	cache.Del(cacheKey)
}
func (n *nodeService) CacheAllocatedStatus(cache *ristretto.Cache[string, any], item *v1.Node) {
	name := item.GetName()
	version := item.GetResourceVersion()
	ns := item.GetNamespace()

	cacheKey := fmt.Sprintf("%s/%s/%s/%s", "NodeAllocatedStatus", ns, name, version)
	_, _ = utils.GetOrSetCache(cache, cacheKey, nodeStatusTTL, func() ([]*kom.ResourceUsageRow, error) {
		tb := kom.DefaultCluster().Name(name).Namespace(ns).WithCache(nodeStatusTTL).Resource(&v1.Node{}).Ctl().Node().ResourceUsageTable()
		return tb, nil
	})

}
func (n *nodeService) RemoveCacheAllocatedStatus(cache *ristretto.Cache[string, any], item *v1.Node) {
	name := item.GetName()
	version := item.GetResourceVersion()
	ns := item.GetNamespace()
	cacheKey := fmt.Sprintf("%s/%s/%s/%s", "NodeAllocatedStatus", ns, name, version)
	cache.Del(cacheKey)
}
