package service

import (
	"fmt"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog/v2"
)

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
	u, _ := utils.GetOrSetCache(cache, cacheKey, ttl, func() (ipUsage, error) {
		total, used, available := kom.DefaultCluster().Name(nodeName).WithCache(ttl).Ctl().Node().IPUsage()
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
	table, _ := utils.GetOrSetCache(cache, cacheKey, ttl, func() ([]*kom.ResourceUsageRow, error) {
		tb := kom.DefaultCluster().Name(name).WithCache(ttl).Ctl().Node().ResourceUsageTable()
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
func (n *nodeService) Watch() error {
	podWatchOnce.Do(func() {
		// watch default 命名空间下 Pod资源 的变更
		var watcher watch.Interface
		var node v1.Node
		cache := kom.DefaultCluster().ClusterCache()

		err := kom.DefaultCluster().Resource(&node).Watch(&watcher).Error
		if err != nil {
			klog.Errorf("PodService Create Watcher Error %v", err)
			return
		}
		go func() {
			klog.V(6).Infof("start watch pod")
			defer watcher.Stop()
			for event := range watcher.ResultChan() {
				err = kom.DefaultCluster().Tools().ConvertRuntimeObjectToTypedObject(event.Object, &node)
				if err != nil {
					klog.V(6).Infof("无法将对象转换为 *v1.Pod 类型: %v", err)
					return
				}
				// 处理事件
				switch event.Type {
				case watch.Added:
					n.CacheAllocatedStatus(cache, &node)
					n.CacheIPUsage(cache, &node)
					klog.V(6).Infof("Added Node [ %s/%s ]\n", node.Namespace, node.Name)
				case watch.Modified:
					n.CacheAllocatedStatus(cache, &node)
					n.CacheIPUsage(cache, &node)
					klog.V(6).Infof("Modified Node [ %s/%s ]\n", node.Namespace, node.Name)
				case watch.Deleted:
					n.RemoveCacheAllocatedStatus(cache, &node)
					n.RemoveCacheIPUsage(cache, &node)
					klog.V(6).Infof("Deleted Node [ %s/%s ]\n", node.Namespace, node.Name)
				}
			}
		}()

	})

	return nil
}
func (n *nodeService) CacheIPUsage(cache *ristretto.Cache[string, any], item *v1.Node) {
	nodeName := item.GetName()
	cacheKey := fmt.Sprintf("%s/%s", "NodeIPUsage", nodeName)
	_, _ = utils.GetOrSetCache(cache, cacheKey, ttl, func() (ipUsage, error) {
		total, used, available := kom.DefaultCluster().Name(nodeName).WithCache(ttl).Ctl().Node().IPUsage()
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
	_, _ = utils.GetOrSetCache(cache, cacheKey, ttl, func() ([]*kom.ResourceUsageRow, error) {
		tb := kom.DefaultCluster().Name(name).Namespace(ns).WithCache(ttl).Resource(&v1.Node{}).Ctl().Node().ResourceUsageTable()
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
