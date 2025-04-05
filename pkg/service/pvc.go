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

type pvcService struct {
	CountList []*pvcCount
	lock      sync.RWMutex
}

// 定义结构体，为按spec.storageclass.name统计数量，包括集群、name、数量
type pvcCount struct {
	ClusterName string // 集群名称
	Name        string // storageClassName
	Count       int    // 数量
}

// IncreasePVCCount 增加pvc统计数据
func (p *pvcService) IncreasePVCCount(selectedCluster string, pvc *corev1.PersistentVolumeClaim) {
	// 从CountList中看是否有该集群、该storageClassName的项，有则加1，无则创建为1
	// 为避免并发操作统计错误，加一个锁
	p.lock.Lock()
	defer p.lock.Unlock()

	// 检查 pvc.Spec.StorageClassName 是否为 nil，避免空指针异常
	if pvc.Spec.StorageClassName == nil {
		return
	}
	h := slice.Filter(p.CountList, func(index int, item *pvcCount) bool {
		return item.ClusterName == selectedCluster && item.Name == *pvc.Spec.StorageClassName
	})
	if len(h) == 0 {
		// 还没有该集群、该storageClassName的项，创建
		p.CountList = append(p.CountList, &pvcCount{
			ClusterName: selectedCluster,
			Name:        *pvc.Spec.StorageClassName,
			Count:       1,
		})
		return
	}
	if len(h) == 1 {
		h[0].Count = h[0].Count + 1
		return
	}

}

// ReducePVCCount 减少pvc统计数据
func (p *pvcService) ReducePVCCount(selectedCluster string, pvc *corev1.PersistentVolumeClaim) {
	// 从CountList中看是否有该集群、该storageClassName的项，有则减1，无则不操作
	// 为避免并发操作统计错误，加一个锁
	p.lock.Lock()
	defer p.lock.Unlock()
	// 检查 pvc.Spec.StorageClassName 是否为 nil，避免空指针异常
	if pvc.Spec.StorageClassName == nil {
		return
	}
	h := slice.Filter(p.CountList, func(index int, item *pvcCount) bool {
		return item.ClusterName == selectedCluster && item.Name == *pvc.Spec.StorageClassName
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

// GetPVCCount 按StorageClassName获取pvc统计数据
func (p *pvcService) GetPVCCount(selectedCluster string, name string) int {
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

func (p *pvcService) Watch() {
	// 设置一个定时器，不断查看是否有集群未开启watch，未开启的话，开启watch
	inst := cron.New()
	_, err := inst.AddFunc("@every 1m", func() {
		// 延迟启动cron
		clusters := ClusterService().ConnectedClusters()
		for _, cluster := range clusters {
			if !cluster.GetClusterWatchStatus("pvc") {
				selectedCluster := ClusterService().ClusterID(cluster)
				watcher := p.watchSingleCluster(selectedCluster)
				cluster.SetClusterWatchStarted("pvc", watcher)
			}
		}
	})
	if err != nil {
		klog.Errorf("新增PVC状态定时更新任务报错: %v\n", err)
	}
	inst.Start()
	klog.V(6).Infof("新增PVC状态定时更新任务【@every 1m】\n")
}

func (p *pvcService) watchSingleCluster(selectedCluster string) watch.Interface {
	// watch  命名空间下 pvc 的变更
	var watcher watch.Interface
	var pvc corev1.PersistentVolumeClaim
	ctx := utils2.GetContextWithAdmin()
	err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&pvc).AllNamespace().Watch(&watcher).Error
	if err != nil {
		klog.Errorf("%s 创建pvc监听器失败 %v", selectedCluster, err)
		return nil
	}
	go func() {
		klog.V(6).Infof("%s start watch pvc", selectedCluster)
		defer watcher.Stop()

		for event := range watcher.ResultChan() {
			err = kom.Cluster(selectedCluster).WithContext(ctx).Tools().ConvertRuntimeObjectToTypedObject(event.Object, &pvc)
			if err != nil {
				klog.V(6).Infof("%s 无法将对象转换为 *v1.PersistentVolumeClaim 类型: %v", selectedCluster, err)
				return
			}
			// 处理事件
			switch event.Type {
			case watch.Added:
				p.IncreasePVCCount(selectedCluster, &pvc)
				klog.V(6).Infof("%s 添加PVC [ %s/%s ]\n", selectedCluster, pvc.Namespace, pvc.Name)
			case watch.Modified:
				// 统计数量，修改跳过
				klog.V(6).Infof("%s 修改PVC [ %s/%s ]\n", selectedCluster, pvc.Namespace, pvc.Name)
			case watch.Deleted:
				p.ReducePVCCount(selectedCluster, &pvc)
				klog.V(6).Infof("%s 删除PVC [ %s/%s ]\n", selectedCluster, pvc.Namespace, pvc.Name)
			}
		}
	}()

	// 延迟设置完成状态，等待 ListWatch完成
	ClusterService().DelayStartFunc(func() {
		ClusterService().SetPVCStatusAggregated(selectedCluster, true)
	})
	return watcher
}
