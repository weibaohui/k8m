package service

import (
	"sync"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/robfig/cron/v3"
	utils2 "github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/kom/kom"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog/v2"
)

type ingressService struct {
	CountList []*ingressCount
	lock      sync.RWMutex
}

// 定义结构体，为按ingressClassName统计数量，包括集群、name、数量
type ingressCount struct {
	ClusterName string // 集群名称
	Name        string // ingressClassName
	Count       int    // 数量
}

// IncreaseIngressCount 增加ingress统计数据
func (p *ingressService) IncreaseIngressCount(selectedCluster string, ingress *networkingv1.Ingress) {
	p.lock.Lock()
	defer p.lock.Unlock()

	// 检查 ingress.Spec.IngressClassName 是否为 nil，避免空指针异常
	if ingress.Spec.IngressClassName == nil {
		return
	}

	h := slice.Filter(p.CountList, func(index int, item *ingressCount) bool {
		return item.ClusterName == selectedCluster && item.Name == *ingress.Spec.IngressClassName
	})
	if len(h) == 0 {
		p.CountList = append(p.CountList, &ingressCount{
			ClusterName: selectedCluster,
			Name:        *ingress.Spec.IngressClassName,
			Count:       1,
		})
		return
	}
	if len(h) == 1 {
		h[0].Count = h[0].Count + 1
		return
	}
}

// ReduceIngressCount 减少ingress统计数据
func (p *ingressService) ReduceIngressCount(selectedCluster string, ingress *networkingv1.Ingress) {
	p.lock.Lock()
	defer p.lock.Unlock()

	// 检查 ingress.Spec.IngressClassName 是否为 nil，避免空指针异常
	if ingress.Spec.IngressClassName == nil {
		return
	}
	h := slice.Filter(p.CountList, func(index int, item *ingressCount) bool {
		return item.ClusterName == selectedCluster && item.Name == *ingress.Spec.IngressClassName
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

// GetIngressCount 按IngressClassName获取ingress统计数据
func (p *ingressService) GetIngressCount(selectedCluster string, name string) int {
	p.lock.RLock()
	defer p.lock.RUnlock()
	for _, item := range p.CountList {
		if item.ClusterName == selectedCluster && item.Name == name {
			return item.Count
		}
	}
	return 0
}

func (p *ingressService) Watch() {
	inst := cron.New()
	_, err := inst.AddFunc("@every 1m", func() {
		clusters := ClusterService().ConnectedClusters()
		for _, cluster := range clusters {
			if !cluster.GetClusterWatchStatus("ingress") {
				selectedCluster := ClusterService().ClusterID(cluster)
				watcher := p.watchSingleCluster(selectedCluster)
				cluster.SetClusterWatchStarted("ingress", watcher)
			}
		}
	})
	if err != nil {
		klog.Errorf("新增Ingress状态定时更新任务报错: %v\n", err)
	}
	inst.Start()
	klog.V(6).Infof("新增Ingress状态定时更新任务【@every 1m】\n")
}

func (p *ingressService) watchSingleCluster(selectedCluster string) watch.Interface {
	var watcher watch.Interface
	var ingress networkingv1.Ingress
	ctx := utils2.GetContextWithAdmin()
	err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&ingress).AllNamespace().Watch(&watcher).Error
	if err != nil {
		klog.Errorf("%s 创建ingress监听器失败 %v", selectedCluster, err)
		return nil
	}
	go func() {
		klog.V(6).Infof("%s start watch ingress", selectedCluster)
		defer watcher.Stop()
		for event := range watcher.ResultChan() {
			err = kom.Cluster(selectedCluster).WithContext(ctx).Tools().ConvertRuntimeObjectToTypedObject(event.Object, &ingress)
			if err != nil {
				klog.V(6).Infof("%s 无法将对象转换为 *v1.Ingress 类型: %v", selectedCluster, err)
				return
			}
			switch event.Type {
			case watch.Added:
				p.IncreaseIngressCount(selectedCluster, &ingress)
				klog.V(6).Infof("%s 添加Ingress [ %s/%s ]\n", selectedCluster, ingress.Namespace, ingress.Name)
			case watch.Modified:
				klog.V(6).Infof("%s 修改Ingress [ %s/%s ]\n", selectedCluster, ingress.Namespace, ingress.Name)
			case watch.Deleted:
				p.ReduceIngressCount(selectedCluster, &ingress)
				klog.V(6).Infof("%s 删除Ingress [ %s/%s ]\n", selectedCluster, ingress.Namespace, ingress.Name)
			}
		}
	}()

	ClusterService().DelayStartFunc(func() {
		ClusterService().SetIngressStatusAggregated(selectedCluster, true)
	})
	return watcher
}
