package service

import (
	"sync"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/robfig/cron/v3"
	utils2 "github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/kom/kom"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog/v2"
)

type pvService struct {
	CountList []*pvCount
	lock      sync.RWMutex
}

// 定义结构体，为按spec.storageclass.name统计数量，包括集群、name、数量
type pvCount struct {
	ClusterName string // 集群名称
	Name        string // storageClassName
	Count       int    // 数量
}

// IncreasePVCount 增加pvc统计数据
func (p *pvService) IncreasePVCount(selectedCluster string, pv *corev1.PersistentVolume) {
	// 从CountList中看是否有该集群、该storageClassName的项，有则加1，无则创建为1
	// 为避免并发操作统计错误，加一个锁
	p.lock.Lock()
	defer p.lock.Unlock()

	h := slice.Filter(p.CountList, func(index int, item *pvCount) bool {
		return item.ClusterName == selectedCluster && item.Name == pv.Spec.StorageClassName
	})
	if len(h) == 0 {
		// 还没有该集群、该storageClassName的项，创建
		p.CountList = append(p.CountList, &pvCount{
			ClusterName: selectedCluster,
			Name:        pv.Spec.StorageClassName,
			Count:       1,
		})
		return
	}
	if len(h) == 1 {
		h[0].Count = h[0].Count + 1
		return
	}

}

// ReducePVCount 减少pv统计数据
func (p *pvService) ReducePVCount(selectedCluster string, pv *corev1.PersistentVolume) {
	// 从CountList中看是否有该集群、该storageClassName的项，有则减1，无则不操作
	// 为避免并发操作统计错误，加一个锁
	p.lock.Lock()
	defer p.lock.Unlock()
	h := slice.Filter(p.CountList, func(index int, item *pvCount) bool {
		return item.ClusterName == selectedCluster && item.Name == pv.Spec.StorageClassName
	})
	if len(h) == 0 {
		return
	}
	if len(h) == 1 {
		h[0].Count = h[0].Count - 1
		if h[0].Count < 0 {
			h[0].Count = 0
		}
	}
}

// GetPVCount 按StorageClassName获取pv统计数据
func (p *pvService) GetPVCount(selectedCluster string, name string) int {
	// 从CountList中看是否有该集群的项，有则返回，无则返回0
	// 为避免并发操作统计错误，加一个锁
	p.lock.RLock()
	defer p.lock.RUnlock()
	for _, item := range p.CountList {
		if item.ClusterName == selectedCluster && item.Name == name {
			return item.Count
		}
	}
	return 0
}

func (p *pvService) Watch() {
	// 设置一个定时器，不断查看是否有集群未开启watch，未开启的话，开启watch
	inst := cron.New()
	_, err := inst.AddFunc("@every 1m", func() {
		// 延迟启动cron
		clusters := ClusterService().ConnectedClusters()
		for _, cluster := range clusters {
			if !cluster.GetClusterWatchStatus("pv") {
				selectedCluster := ClusterService().ClusterID(cluster)
				watcher := p.watchSingleCluster(selectedCluster)
				cluster.SetClusterWatchStarted("pv", watcher)
			}
		}
	})
	if err != nil {
		klog.Errorf("新增PV状态定时更新任务报错: %v\n", err)
	}
	inst.Start()
	klog.V(6).Infof("新增PV状态定时更新任务【@every 1m】\n")
}

func (p *pvService) watchSingleCluster(selectedCluster string) watch.Interface {
	// watch  命名空间下 pv 的变更
	var watcher watch.Interface
	var pv corev1.PersistentVolume
	ctx := utils2.GetContextWithAdmin()
	err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&pv).AllNamespace().Watch(&watcher).Error
	if err != nil {
		klog.Errorf("%s 创建pv监听器失败 %v", selectedCluster, err)
		return nil
	}
	go func() {
		klog.V(6).Infof("%s start watch pv", selectedCluster)
		defer watcher.Stop()
		for event := range watcher.ResultChan() {
			err = kom.Cluster(selectedCluster).WithContext(ctx).Tools().ConvertRuntimeObjectToTypedObject(event.Object, &pv)
			if err != nil {
				klog.V(6).Infof("%s 无法将对象转换为 *v1.PersistentVolume 类型: %v", selectedCluster, err)
				return
			}
			// 处理事件
			switch event.Type {
			case watch.Added:
				p.IncreasePVCount(selectedCluster, &pv)
				klog.V(6).Infof("%s 添加PV [ %s/%s ]\n", selectedCluster, pv.Namespace, pv.Name)
			case watch.Modified:
				// 统计数量，修改跳过
				klog.V(6).Infof("%s 修改PV [ %s/%s ]\n", selectedCluster, pv.Namespace, pv.Name)
			case watch.Deleted:
				p.ReducePVCount(selectedCluster, &pv)
				klog.V(6).Infof("%s 删除PV [ %s/%s ]\n", selectedCluster, pv.Namespace, pv.Name)
			}
		}
	}()

	// 延迟设置完成状态，等待 ListWatch完成
	ClusterService().DelayStartFunc(func() {
		ClusterService().SetPVStatusAggregated(selectedCluster, true)
	})
	return watcher
}
