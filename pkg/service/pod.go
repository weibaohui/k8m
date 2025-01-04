package service

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog/v2"
)

var podWatchOnce sync.Once
var ttl = 24 * time.Hour

type podService struct {
}

func (p *podService) StreamPodLogs(ctx context.Context, selectedCluster string, ns, name string, logOptions *v1.PodLogOptions) (io.ReadCloser, error) {

	// 检查logOptions
	//  at most one of `sinceTime` or `sinceSeconds` may be specified
	if (logOptions.SinceTime != nil) && (logOptions.SinceSeconds != nil && *logOptions.SinceSeconds > 0) {
		// 同时设置，保留SinceSeconds
		logOptions.SinceTime = nil
	}
	if logOptions.SinceSeconds != nil && *logOptions.SinceSeconds == 0 {
		logOptions.SinceSeconds = nil
	}
	var stream io.ReadCloser
	err := kom.Cluster(selectedCluster).WithContext(ctx).Namespace(ns).Name(name).ContainerName(logOptions.Container).GetLogs(&stream, logOptions).Error

	return stream, err
}

// SetAllocatedStatus 设置节点的分配状态
// pod 资源状态一般不会变化，变化了version也会变
func (p *podService) SetAllocatedStatus(selectedCluster string, cache *ristretto.Cache[string, any], item unstructured.Unstructured) unstructured.Unstructured {
	podName := item.GetName()
	version := item.GetResourceVersion()
	ns := item.GetNamespace()
	cacheKey := fmt.Sprintf("%s/%s/%s/%s", "PodAllocatedStatus", ns, podName, version)
	table, err := utils.GetOrSetCache(cache, cacheKey, ttl, func() ([]*kom.ResourceUsageRow, error) {
		tb := kom.Cluster(selectedCluster).Name(podName).Namespace(ns).WithCache(ttl).Resource(&v1.Pod{}).Ctl().Pod().ResourceUsageTable()
		return tb, nil
	})
	if err != nil {
		return item
	}

	for _, row := range table {
		if row.ResourceType == "cpu" {
			utils.AddOrUpdateAnnotations(&item, map[string]string{
				"cpu.request":         fmt.Sprintf("%s", row.Request),
				"cpu.requestFraction": fmt.Sprintf("%s", row.RequestFraction),
				"cpu.limit":           fmt.Sprintf("%s", row.Limit),
				"cpu.limitFraction":   fmt.Sprintf("%s", row.LimitFraction),
				"cpu.total":           fmt.Sprintf("%s", row.Total),
			})
		} else if row.ResourceType == "memory" {
			utils.AddOrUpdateAnnotations(&item, map[string]string{
				"memory.request":         fmt.Sprintf("%s", row.Request),
				"memory.requestFraction": fmt.Sprintf("%s", row.RequestFraction),
				"memory.limit":           fmt.Sprintf("%s", row.Limit),
				"memory.limitFraction":   fmt.Sprintf("%s", row.LimitFraction),
				"memory.total":           fmt.Sprintf("%s", row.Total),
			})
		}
	}
	return item
}
func (p *podService) CacheAllocatedStatus(selectedCluster string, cache *ristretto.Cache[string, any], item *v1.Pod) {
	podName := item.GetName()
	version := item.GetResourceVersion()
	ns := item.GetNamespace()
	cacheKey := fmt.Sprintf("%s/%s/%s/%s", "PodAllocatedStatus", ns, podName, version)
	_, _ = utils.GetOrSetCache(cache, cacheKey, ttl, func() ([]*kom.ResourceUsageRow, error) {
		tb := kom.Cluster(selectedCluster).Name(podName).Namespace(ns).WithCache(ttl).Resource(&v1.Pod{}).Ctl().Pod().ResourceUsageTable()
		return tb, nil
	})

}
func (p *podService) RemoveCacheAllocatedStatus(selectedCluster string, cache *ristretto.Cache[string, any], item *v1.Pod) {
	podName := item.GetName()
	version := item.GetResourceVersion()
	ns := item.GetNamespace()
	cacheKey := fmt.Sprintf("%s/%s/%s/%s", "PodAllocatedStatus", ns, podName, version)
	cache.Del(cacheKey)
}
func (p *podService) Watch() error {
	podWatchOnce.Do(func() {
		clusters := ClusterService().ConnectedClusters()
		for _, cluster := range clusters {
			selectedCluster := fmt.Sprintf("%s/%s", cluster.FileName, cluster.ContextName)
			p.watchSingleCluster(selectedCluster)
		}
	})

	return nil
}

func (p *podService) watchSingleCluster(selectedCluster string) {
	// watch default 命名空间下 Pod资源 的变更
	var watcher watch.Interface
	var pod v1.Pod
	err := kom.Cluster(selectedCluster).Resource(&pod).Namespace(v1.NamespaceAll).Watch(&watcher).Error
	if err != nil {
		klog.Errorf("PodService Create Watcher Error %v", err)
		return
	}
	cache := kom.Cluster(selectedCluster).ClusterCache()
	go func() {
		klog.V(6).Infof("start watch pod")
		defer watcher.Stop()
		for event := range watcher.ResultChan() {
			err := kom.Cluster(selectedCluster).Tools().ConvertRuntimeObjectToTypedObject(event.Object, &pod)
			if err != nil {
				klog.V(6).Infof("无法将对象转换为 *v1.Pod 类型: %v", err)
				return
			}
			// 处理事件
			switch event.Type {
			case watch.Added:
				p.CacheAllocatedStatus(selectedCluster, cache, &pod)
				fmt.Printf("Added Pod [ %s/%s ]\n", pod.Namespace, pod.Name)
			case watch.Modified:
				p.CacheAllocatedStatus(selectedCluster, cache, &pod)
				fmt.Printf("Modified Pod [ %s/%s ]\n", pod.Namespace, pod.Name)
			case watch.Deleted:
				p.RemoveCacheAllocatedStatus(selectedCluster, cache, &pod)
				fmt.Printf("Deleted Pod [ %s/%s ]\n", pod.Namespace, pod.Name)
			}
		}
	}()
}
