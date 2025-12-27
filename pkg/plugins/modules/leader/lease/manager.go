package lease

import (
	"context"
	"crypto/sha1"
	"fmt"
	"strings"
	"time"

	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/flag"
	coordinationv1 "k8s.io/api/coordination/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
)

// Manager 中文函数注释：Lease 管理器接口，负责创建/续约/删除以及监听与清理。
type Manager interface {
	Init(ctx context.Context, opts Options) error
	EnsureOnConnect(ctx context.Context, clusterID string) error
	EnsureOnDisconnect(ctx context.Context, clusterID string) error
	StartWatcher(ctx context.Context, onConnect func(string), onDisconnect func(string)) error
	StartLeaderCleanup(ctx context.Context) error
}

type manager struct {
	clientset    *kubernetes.Clientset
	namespace    string
	instanceID   string
	durationSec  int
	renewSec     int
	resyncPeriod time.Duration
}

// NewManager 中文函数注释：创建一个 Lease 管理器实例。
func NewManager() Manager { return &manager{} }

// Init 中文函数注释：初始化 Lease 管理器，设置宿主 ClientSet、命名空间与续约参数。
func (m *manager) Init(ctx context.Context, opts Options) error {
	klog.V(6).Infof("Init lease")
	cs, hasCluster, err := utils.GetClientSet(opts.ClusterID)
	if err != nil {
		klog.V(6).Infof("GetClientSet %v", err.Error())
		return fmt.Errorf("初始化宿主 GetClientSet 失败: %w", err)
	}
	// 没有可用的集群，那么就无法执行这个模式了
	if !hasCluster {
		klog.V(2).Infof("[Lease] 无可用的 K8s 集群,退出初始化")
		return fmt.Errorf("no available k8s cluster")
	}

	m.clientset = cs
	if opts.Namespace == "" {
		m.namespace = utils.DetectNamespace()
	} else {
		m.namespace = opts.Namespace
	}
	m.durationSec = opts.LeaseDurationSeconds
	if m.durationSec <= 0 {
		m.durationSec = 60
	}
	m.renewSec = opts.LeaseRenewIntervalSeconds
	if m.renewSec <= 0 || m.renewSec >= m.durationSec {
		m.renewSec = m.durationSec / 3
		if m.renewSec <= 0 {
			m.renewSec = 20
		}
	}
	if opts.ResyncPeriod <= 0 {
		m.resyncPeriod = 30 * time.Second
	} else {
		m.resyncPeriod = opts.ResyncPeriod
	}
	m.instanceID = utils.GenerateInstanceID()
	klog.V(6).Infof("Lease 管理器初始化完成，ns=%s, duration=%ds, renew=%ds, instance=%s", m.namespace, m.durationSec, m.renewSec, m.instanceID)
	return nil
}

// EnsureOnConnect 中文函数注释：在连接某个集群前确保该集群的 Lease 存在并进入续约循环；若已存在则仅提示并返回。
func (m *manager) EnsureOnConnect(ctx context.Context, clusterID string) error {
	if m.clientset == nil {
		return nil
	}

	lc := m.clientset.CoordinationV1().Leases(m.namespace)
	name := m.leaseName(clusterID)
	_, err := lc.Get(ctx, name, metav1.GetOptions{})
	if err == nil {
		klog.V(6).Infof("Lease 已存在，跳过创建：%s", name)
		return nil
	}
	labels := map[string]string{
		"app":       "k8m",
		"type":      "cluster-sync",
		"clusterID": utils.UrlSafeBase64Encode(clusterID),
	}

	m.createOrUpdateLease(ctx, name, labels)
	go m.loopRenewLease(name, labels, m.durationSec, m.renewSec, m.instanceID)
	return nil
}

// EnsureOnDisconnect 中文函数注释：在断开某个集群前确保删除对应的 Lease；若不存在则跳过。
func (m *manager) EnsureOnDisconnect(ctx context.Context, clusterID string) error {
	if m.clientset == nil {
		return nil
	}
	lc := m.clientset.CoordinationV1().Leases(m.namespace)
	name := m.leaseName(clusterID)
	_ = lc.Delete(ctx, name, metav1.DeleteOptions{})
	klog.V(6).Infof("删除 Lease：%s", name)
	return nil
}

// StartWatcher 中文函数注释：启动 Lease 监听器，对有效 Lease 的新增/更新触发本地连接，对删除触发本地断开。
func (m *manager) StartWatcher(ctx context.Context, onConnect func(string), onDisconnect func(string)) error {
	if m.clientset == nil {
		return nil
	}
	// 仅监听指定命名空间和标签
	selector := labels.SelectorFromSet(labels.Set{"app": "k8m", "type": "cluster-sync"})
	factory := informers.NewSharedInformerFactoryWithOptions(m.clientset, m.resyncPeriod,
		informers.WithNamespace(m.namespace), informers.WithTweakListOptions(func(lo *metav1.ListOptions) {
			lo.LabelSelector = selector.String()
		}))
	informer := factory.Coordination().V1().Leases().Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			l := obj.(*coordinationv1.Lease)
			if !isLeaseValid(l, m.durationSec) {
				return
			}
			cid := l.Labels["clusterID"]
			clusterID, err := utils.UrlSafeBase64Decode(cid)
			if err != nil {
				klog.V(6).Infof("解码 Lease 标签 clusterID 失败：%v", err)
				return
			}
			if deref(l.Spec.HolderIdentity) == m.instanceID {
				return
			}
			klog.V(6).Infof("有效 Lease 新增，外部连接集群：%s", clusterID)
			go onConnect(clusterID)
		},

		DeleteFunc: func(obj any) {
			var l *coordinationv1.Lease
			if dfo, ok := obj.(cache.DeletedFinalStateUnknown); ok {
				l, _ = dfo.Obj.(*coordinationv1.Lease)
			} else {
				l, _ = obj.(*coordinationv1.Lease)
			}
			if l == nil {
				return
			}

			cid := l.Labels["clusterID"]
			clusterID, err := utils.UrlSafeBase64Decode(cid)
			if err != nil {
				klog.V(6).Infof("解码 Lease 标签 clusterID 失败：%v", err)
				return
			}

			klog.V(6).Infof("Lease 删除，断开本地集群：%s", clusterID)
			go onDisconnect(clusterID)
		},
	})

	factory.Start(ctx.Done())
	klog.V(6).Infof("Lease 监听器已启动，命名空间：%s", m.namespace)
	return nil
}

// StartLeaderCleanup 仅由 Leader 调用，定期清理过期 Lease；通过 Delete 事件驱动所有实例断开。
func (m *manager) StartLeaderCleanup(ctx context.Context) error {
	if m.clientset == nil {
		return nil
	}
	ticker := time.NewTicker(30 * time.Second)
	go func() {
		for {
			select {
			case <-ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				lc := m.clientset.CoordinationV1().Leases(m.namespace)
				ls, err := lc.List(ctx, metav1.ListOptions{LabelSelector: "app=k8m,type=cluster-sync"})
				if err != nil {
					klog.V(6).Infof("清理过期 Lease 失败：%v", err)
					continue
				}
				for _, l := range ls.Items {
					if !isLeaseValid(&l, m.durationSec) {
						_ = lc.Delete(ctx, l.Name, metav1.DeleteOptions{})
						klog.V(6).Infof("删除过期 Lease：%s", l.Name)
					}
				}
			}
		}
	}()
	klog.V(6).Infof("Lease 过期清理任务启动（Leader）")
	return nil
}

// createOrUpdateLease 中文函数注释：创建 Lease；若已存在则更新其续约时间与持有者。
func (m *manager) createOrUpdateLease(ctx context.Context, name string, labels map[string]string) {
	lc := m.clientset.CoordinationV1().Leases(m.namespace)
	n := time.Now()
	ld := int32(m.durationSec)
	l := &coordinationv1.Lease{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Spec: coordinationv1.LeaseSpec{
			HolderIdentity:       ptrString(m.instanceID),
			LeaseDurationSeconds: &ld,
			RenewTime:            &metav1.MicroTime{Time: n},
			LeaseTransitions:     ptrInt32(1),
		},
	}
	_, err := lc.Create(ctx, l, metav1.CreateOptions{})
	if err != nil && strings.Contains(err.Error(), "already exists") {
		// 已存在则更新
		_, _ = lc.Update(ctx, l, metav1.UpdateOptions{})
	}
}

// loopRenewLease 中文函数注释：循环续约 Lease，更新续约时间与持有者信息。
func (m *manager) loopRenewLease(name string, labels map[string]string, durationSec int, renewSec int, instanceID string) {
	lc := m.clientset.CoordinationV1().Leases(m.namespace)
	t := time.NewTicker(time.Duration(renewSec) * time.Second)
	defer t.Stop()
	for range t.C {
		n := time.Now()
		ld := int32(durationSec)
		l := &coordinationv1.Lease{
			ObjectMeta: metav1.ObjectMeta{
				Name:   name,
				Labels: labels,
			},
			Spec: coordinationv1.LeaseSpec{
				HolderIdentity:       ptrString(instanceID),
				LeaseDurationSeconds: &ld,
				RenewTime:            &metav1.MicroTime{Time: n},
			},
		}
		_, err := lc.Update(context.Background(), l, metav1.UpdateOptions{})
		if err != nil {
			klog.V(6).Infof("续约 Lease 失败：%v", err)
		}
	}
}

// leaseName 中文函数注释：生成 Lease 名称，格式：<product>-cluster-<sha1(clusterID)前4字节>
func (m *manager) leaseName(clusterID string) string {
	s := sha1.Sum([]byte(clusterID))
	first4 := s[:4]
	product := flag.Init().ProductName
	return fmt.Sprintf("%s-cluster-%x", product, first4)
}

// isLeaseValid 中文函数注释：判断 Lease 是否仍在有效期内。
func isLeaseValid(l *coordinationv1.Lease, durationSec int) bool {
	if l == nil || l.Spec.RenewTime == nil || l.Spec.LeaseDurationSeconds == nil {
		return false
	}
	d := time.Duration(durationSec) * time.Second
	return time.Since(l.Spec.RenewTime.Time) < d
}

func ptrString(s string) *string { return &s }
func ptrInt32(i int32) *int32    { return &i }
func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
