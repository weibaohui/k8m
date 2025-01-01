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
	"k8s.io/klog/v2"
)

var podOnce sync.Once

type podService struct {
	Cache *ristretto.Cache[string, any]
}

func newPodService() *podService {
	podOnce.Do(func() {
		klog.V(6).Infof("init localPodService")
		cache, err := ristretto.NewCache(&ristretto.Config[string, any]{
			NumCounters: 1e7,     // number of keys to track frequency of (10M).
			MaxCost:     1 << 30, // maximum cost of cache (1GB).
			BufferItems: 64,      // number of keys per Get buffer.
		})
		if err != nil {
			klog.Errorf("Failed to create cache: %v", err)
		}
		localPodService = &podService{Cache: cache}
	})

	return localPodService
}

func (p *podService) StreamPodLogs(ctx context.Context, ns, name string, logOptions *v1.PodLogOptions) (io.ReadCloser, error) {

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
	err := kom.DefaultCluster().WithContext(ctx).Namespace(ns).Name(name).ContainerName(logOptions.Container).GetLogs(&stream, logOptions).Error

	return stream, err
}

// SetAllocatedStatus 设置节点的分配状态
// pod 资源状态一般不会变化，变化了version也会变
func (p *podService) SetAllocatedStatus(item unstructured.Unstructured) unstructured.Unstructured {
	podName := item.GetName()
	version := item.GetResourceVersion()
	ns := item.GetNamespace()
	cacheKey := fmt.Sprintf("%s/%s/%s", ns, podName, version)
	table, err := utils.GetOrSetCache(p.Cache, cacheKey, 24*time.Hour, func() ([]*kom.ResourceUsageRow, error) {
		tb := kom.DefaultCluster().Name(podName).Namespace(ns).Resource(&v1.Pod{}).Ctl().Pod().ResourceUsageTable()
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
