package service

import (
	"fmt"
	"time"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/robfig/cron/v3"
	utils2 "github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/flag"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/utils"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog/v2"
)

// StatusCount 定义结构体，以namespace为key进行统计，累计所有的pod数量，以及cpu、内存用量
type StatusCount struct {
	ClusterName    string // 集群名称
	Namespace      string
	PodCount       int     // 数量
	CPURequest     float64 // cpu请求量
	CPULimit       float64 // cpu上限量
	CPURealtime    float64 // cpu实时量
	MemoryRequest  float64 // 内存请求量
	MemoryLimit    float64 // 内存上限量
	MemoryRealtime float64 // 内存实时量
}

func (p *podService) getTTL() time.Duration {
	cfg := flag.Init()
	if cfg.ResourceCacheTimeout > 0 {
		return time.Duration(cfg.ResourceCacheTimeout) * time.Second
	}
	return 1 * time.Minute
}
func (p *podService) IncreasePodCount(selectedCluster string, pod *corev1.Pod) {
	// 从CountList中看是否有该集群、该namespace的项，有则加1，无则创建为1
	// 为避免并发操作统计错误，加一个锁
	p.lock.Lock()
	defer p.lock.Unlock()

	h := slice.Filter(p.CountList, func(index int, item *StatusCount) bool {
		return item.ClusterName == selectedCluster && item.Namespace == pod.Namespace
	})
	ctx := utils2.GetContextWithAdmin()
	cacheKey := fmt.Sprintf("%s/%s/%s/%s", "PodResourceUsage", pod.Namespace, pod.Name, pod.ResourceVersion)
	table, err := utils.GetOrSetCache(kom.Cluster(selectedCluster).ClusterCache(), cacheKey, p.getTTL(), func() (*kom.ResourceUsageResult, error) {
		tb, err := kom.Cluster(selectedCluster).WithContext(ctx).Name(pod.Name).Namespace(pod.Namespace).Resource(&v1.Pod{}).Ctl().Pod().ResourceUsage()
		return tb, err
	})
	if err != nil {
		klog.V(6).Infof("get pod resource error %s/%s: %v", pod.Namespace, pod.Name, err)
		return
	}

	if len(h) == 0 {
		// 还没有该集群、该namespace的项，创建
		sc := &StatusCount{
			ClusterName: selectedCluster,
			Namespace:   pod.Namespace,
			PodCount:    1,
		}
		if table != nil {
			for name, quantity := range table.Requests {
				switch name {
				case "cpu":
					sc.CPURequest += quantity.AsApproximateFloat64()
				case "memory":
					sc.MemoryRequest += quantity.AsApproximateFloat64()
				}
			}
			for name, quantity := range table.Limits {
				switch name {
				case "cpu":
					sc.CPULimit += quantity.AsApproximateFloat64()
				case "memory":
					sc.MemoryLimit += quantity.AsApproximateFloat64()
				}
			}
			for name, quantity := range table.Limits {
				switch name {
				case "cpu":
					sc.CPURealtime += quantity.AsApproximateFloat64()
				case "memory":
					sc.MemoryRealtime += quantity.AsApproximateFloat64()
				}
			}
		}

		p.CountList = append(p.CountList, sc)
		return
	}
	if len(h) == 1 {
		sc := h[0]
		sc.PodCount += 1
		if table != nil {
			for name, quantity := range table.Requests {
				switch name {
				case "cpu":
					sc.CPURequest += quantity.AsApproximateFloat64()
				case "memory":
					sc.MemoryRequest += quantity.AsApproximateFloat64()
				}
			}
			for name, quantity := range table.Limits {
				switch name {
				case "cpu":
					sc.CPULimit += quantity.AsApproximateFloat64()
				case "memory":
					sc.MemoryLimit += quantity.AsApproximateFloat64()
				}
			}
			for name, quantity := range table.Realtime {
				switch name {
				case "cpu":
					sc.CPURealtime += quantity.AsApproximateFloat64()
				case "memory":
					sc.MemoryRealtime += quantity.AsApproximateFloat64()
				}
			}
		}
		return
	}

}

func (p *podService) ReducePodCount(selectedCluster string, pod *corev1.Pod) {
	// 从CountList中看是否有该集群、该namespace的项，有则减1，无则不操作
	// 为避免并发操作统计错误，加一个锁
	p.lock.Lock()
	defer p.lock.Unlock()
	h := slice.Filter(p.CountList, func(index int, item *StatusCount) bool {
		return item.ClusterName == selectedCluster && item.Namespace == pod.Namespace
	})
	if len(h) == 0 {
		return
	}
	if len(h) == 1 {
		ctx := utils2.GetContextWithAdmin()
		cacheKey := fmt.Sprintf("%s/%s/%s/%s", "PodResourceUsage", pod.Namespace, pod.Name, pod.ResourceVersion)
		table, err := utils.GetOrSetCache(kom.Cluster(selectedCluster).ClusterCache(), cacheKey, p.getTTL(), func() (*kom.ResourceUsageResult, error) {
			tb, err := kom.Cluster(selectedCluster).WithContext(ctx).Name(pod.Name).Namespace(pod.Namespace).Resource(&v1.Pod{}).Ctl().Pod().ResourceUsage()
			return tb, err
		})
		if err != nil {
			klog.V(6).Infof("get pod resource error %s/%s: %v", pod.Namespace, pod.Name, err)
			return
		}

		sc := h[0]
		sc.PodCount -= 1
		if sc.PodCount < 0 {
			sc.PodCount = 0
		}
		if table == nil {
			return
		}
		for name, quantity := range table.Requests {
			switch name {
			case "cpu":
				sc.CPURequest -= quantity.AsApproximateFloat64()
			case "memory":
				sc.MemoryRequest -= quantity.AsApproximateFloat64()
			}
		}
		for name, quantity := range table.Limits {
			switch name {
			case "cpu":
				sc.CPULimit -= quantity.AsApproximateFloat64()
			case "memory":
				sc.MemoryLimit -= quantity.AsApproximateFloat64()
			}
		}
		for name, quantity := range table.Realtime {
			switch name {
			case "cpu":
				sc.CPURealtime -= quantity.AsApproximateFloat64()
			case "memory":
				sc.MemoryRealtime -= quantity.AsApproximateFloat64()
			}
		}
	}
}

// SetAllocatedStatusOnPod 设置节点的分配状态
// pod 资源状态一般不会变化，变化了version也会变
func (p *podService) SetAllocatedStatusOnPod(selectedCluster string, item unstructured.Unstructured) unstructured.Unstructured {
	podName := item.GetName()
	version := item.GetResourceVersion()
	ns := item.GetNamespace()
	ctx := utils2.GetContextWithAdmin()
	cacheKey := fmt.Sprintf("%s/%s/%s/%s", "PodAllocatedStatus", ns, podName, version)
	table, err := utils.GetOrSetCache(kom.Cluster(selectedCluster).ClusterCache(), cacheKey, p.getTTL(), func() ([]*kom.ResourceUsageRow, error) {
		tb, err := kom.Cluster(selectedCluster).WithContext(ctx).Name(podName).Namespace(ns).Resource(&v1.Pod{}).Ctl().Pod().ResourceUsageTable()
		return tb, err
	})
	if err != nil {
		return item
	}

	for _, row := range table {
		if row.ResourceType == "cpu" {
			utils.AddOrUpdateAnnotations(&item, map[string]string{
				"cpu.request":          row.Request,
				"cpu.requestFraction":  row.RequestFraction,
				"cpu.limit":            row.Limit,
				"cpu.limitFraction":    row.LimitFraction,
				"cpu.realtime":         row.Realtime,
				"cpu.realtimeFraction": row.RealtimeFraction,
				"cpu.total":            row.Total,
			})
		} else if row.ResourceType == "memory" {
			utils.AddOrUpdateAnnotations(&item, map[string]string{
				"memory.request":          row.Request,
				"memory.requestFraction":  row.RequestFraction,
				"memory.limit":            row.Limit,
				"memory.limitFraction":    row.LimitFraction,
				"memory.realtime":         row.Realtime,
				"memory.realtimeFraction": row.RealtimeFraction,
				"memory.total":            row.Total,
			})
		}
	}
	return item
}

func (p *podService) SetStatusCountOnNamespace(selectedCluster string, item unstructured.Unstructured) unstructured.Unstructured {
	p.lock.RLock()
	defer p.lock.RUnlock()
	h := slice.Filter(p.CountList, func(index int, cc *StatusCount) bool {
		return cc.ClusterName == selectedCluster && cc.Namespace == item.GetName()
	})
	if len(h) == 1 {
		sc := h[0]
		utils.AddOrUpdateAnnotations(&item, map[string]string{
			"cpu.request":     fmt.Sprintf("%.2f", sc.CPURequest),
			"cpu.limit":       fmt.Sprintf("%.2f", sc.CPULimit),
			"cpu.realtime":    fmt.Sprintf("%.2f", sc.CPURealtime),
			"memory.request":  fmt.Sprintf("%.2f", sc.MemoryRequest/(1024*1024*1024)),
			"memory.limit":    fmt.Sprintf("%.2f", sc.MemoryLimit/(1024*1024*1024)),
			"memory.realtime": fmt.Sprintf("%.2f", sc.MemoryRealtime/(1024*1024*1024)),
			"pod.count.total": fmt.Sprintf("%d", sc.PodCount),
		})
	}

	return item
}

// CacheKey 获取缓存key
func (p *podService) CacheKey(item *v1.Pod) string {
	podName := item.GetName()
	version := item.GetResourceVersion()
	ns := item.GetNamespace()
	cacheKey := fmt.Sprintf("%s/%s/%s/%s", "PodAllocatedStatus", ns, podName, version)
	return cacheKey

}
func (p *podService) CacheAllocatedStatus(selectedCluster string, item *v1.Pod) {
	podName := item.GetName()
	ns := item.GetNamespace()
	cacheKey := p.CacheKey(item)
	ctx := utils2.GetContextWithAdmin()
	_, _ = utils.GetOrSetCache(kom.Cluster(selectedCluster).ClusterCache(), cacheKey, p.getTTL(), func() ([]*kom.ResourceUsageRow, error) {
		tb, err := kom.Cluster(selectedCluster).WithContext(ctx).Name(podName).Namespace(ns).Resource(&v1.Pod{}).Ctl().Pod().ResourceUsageTable()
		return tb, err
	})

}
func (p *podService) RemoveCacheAllocatedStatus(selectedCluster string, item *v1.Pod) {
	cacheKey := p.CacheKey(item)
	kom.Cluster(selectedCluster).ClusterCache().Del(cacheKey)
}

func (p *podService) Watch() {
	// 设置一个定时器，不断查看是否有集群未开启watch，未开启的话，开启watch
	inst := cron.New()
	_, err := inst.AddFunc("@every 1m", func() {
		// 延迟启动cron
		clusters := ClusterService().ConnectedClusters()
		for _, cluster := range clusters {
			if !cluster.GetClusterWatchStatus("pod") {
				selectedCluster := ClusterService().ClusterID(cluster)
				watcher := p.watchSingleCluster(selectedCluster)
				cluster.SetClusterWatchStarted("pod", watcher)
			}
		}
	})
	if err != nil {
		klog.Errorf("新增Pod状态定时更新任务报错: %v\n", err)
	}
	inst.Start()
	klog.V(6).Infof("新增Pod状态定时更新任务【@every 1m】\n")
}

func (p *podService) watchSingleCluster(selectedCluster string) watch.Interface {
	// watch default 命名空间下 Pod资源 的变更
	ctx := utils2.GetContextWithAdmin()

	var watcher watch.Interface
	var pod v1.Pod
	err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&pod).AllNamespace().Watch(&watcher).Error
	if err != nil {
		klog.Errorf("%s 创建Pod监听器失败 %v", selectedCluster, err)
		return nil
	}
	go func() {
		klog.V(6).Infof("%s start watch pod", selectedCluster)
		defer watcher.Stop()
		for event := range watcher.ResultChan() {
			err = kom.Cluster(selectedCluster).WithContext(ctx).Tools().ConvertRuntimeObjectToTypedObject(event.Object, &pod)
			if err != nil {
				klog.V(6).Infof("%s 无法将对象转换为 *v1.Pod 类型: %v", selectedCluster, err)
				return
			}
			// 处理事件
			switch event.Type {
			case watch.Added:
				p.CacheAllocatedStatus(selectedCluster, &pod)
				p.IncreasePodCount(selectedCluster, &pod)
				// 新增Pod时，保存Pod标签
				p.UpdatePodLabels(selectedCluster, pod.Namespace, pod.Name, pod.Labels)
				klog.V(6).Infof("%s 添加Pod [ %s/%s ] 标签数量: %d\n", selectedCluster, pod.Namespace, pod.Name, len(pod.Labels))
			case watch.Modified:
				p.RemoveCacheAllocatedStatus(selectedCluster, &pod)
				p.CacheAllocatedStatus(selectedCluster, &pod)
				// 修改Pod时，更新Pod标签
				p.UpdatePodLabels(selectedCluster, pod.Namespace, pod.Name, pod.Labels)
				klog.V(6).Infof("%s 修改Pod [ %s/%s ] 标签数量: %d\n", selectedCluster, pod.Namespace, pod.Name, len(pod.Labels))
			case watch.Deleted:
				p.RemoveCacheAllocatedStatus(selectedCluster, &pod)
				p.ReducePodCount(selectedCluster, &pod)
				// 删除Pod时，删除Pod标签
				p.DeletePodLabels(selectedCluster, pod.Namespace, pod.Name)
				klog.V(6).Infof("%s 删除Pod [ %s/%s ]\n", selectedCluster, pod.Namespace, pod.Name)
			}
		}
	}()

	// 延迟设置完成状态，等待Pod ListWatch完成
	ClusterService().DelayStartFunc(func() {
		ClusterService().SetPodStatusAggregated(selectedCluster, true)
	})
	return watcher
}
