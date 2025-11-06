package leader

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/service"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/klog/v2"
)

// Config Leader 选举配置
// 包含命名空间、锁名称、选举时长、续租截止、重试周期以及领导开始/结束回调。
// LeaseDuration > RenewDeadline > RetryPeriod * 3
type Config struct {
	Namespace        string
	LockName         string
	LeaseDuration    time.Duration // 租约持续时间，默认 15s
	RenewDeadline    time.Duration // 续租截止时间，默认 10s
	RetryPeriod      time.Duration // 重试周期，默认 2s
	ClusterID        string        // ClusterID 指定的集群唯一ID（文件名/Context），优先使用该集群的配置
	OnStartedLeading func(ctx context.Context)
	OnStoppedLeading func()
}

// Run 启动 Leader 选举逻辑
// 支持集群优先级：指定 ClusterID -> InCluster -> 本地 kubeconfig。
// 仅当以上方式均不可用时，降级为本地 Leader（不进行选举）。
func Run(ctx context.Context, cfg Config) error {
	clientset, hasCluster, err := getClientset(cfg.ClusterID)
	if err != nil {
		return fmt.Errorf("get clientset failed: %w", err)
	}

	if cfg.Namespace == "" {
		cfg.Namespace = detectNamespace()
	}

	// 非集群模式：既没有指定集群可用，也不是 InCluster，也没有本地 kubeconfig
	if !hasCluster {
		klog.V(2).Infof("[leader] 无可用的 K8s 集群，直接作为 Leader 运行（不进行选举）")
		if cfg.OnStartedLeading != nil {
			cfg.OnStartedLeading(ctx)
		}
		return nil
	}

	if cfg.LeaseDuration == 0 {
		cfg.LeaseDuration = 15 * time.Second
	}
	if cfg.RenewDeadline == 0 {
		cfg.RenewDeadline = 10 * time.Second
	}
	if cfg.RetryPeriod == 0 {
		cfg.RetryPeriod = 2 * time.Second
	}

	id, _ := os.Hostname()
	id = fmt.Sprintf("%s-%s", id, utils.RandNLengthString(3))
	klog.V(2).Infof("[leader] 我的选举 ID：%s", id)
	lock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      cfg.LockName,
			Namespace: cfg.Namespace,
		},
		Client: clientset.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: id,
		},
	}

	leaderelectionCfg := leaderelection.LeaderElectionConfig{
		Lock:            lock,
		ReleaseOnCancel: true,
		LeaseDuration:   cfg.LeaseDuration,
		RenewDeadline:   cfg.RenewDeadline,
		RetryPeriod:     cfg.RetryPeriod,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(c context.Context) {
				if cfg.OnStartedLeading != nil {
					cfg.OnStartedLeading(c)
				}
			},
			OnStoppedLeading: func() {
				if cfg.OnStoppedLeading != nil {
					cfg.OnStoppedLeading()
				}
			},
			OnNewLeader: func(identity string) {
				if identity == id {
					klog.V(2).Infof("[leader] 我成为新的 Leader：%s", id)
				} else {
					klog.V(2).Infof("[leader] 选举产生新的 Leader：%s", identity)
				}
			},
		},
	}

	klog.V(2).Infof("[leader] 开始进行 Leader 选举（锁=%s/%s）", cfg.Namespace, cfg.LockName)
	leaderelection.RunOrDie(ctx, leaderelectionCfg)
	return nil
}

// getClientset 获取 Kubernetes ClientSet
// 优先顺序：
// 1. 指定 ClusterID（外部集群）
// 2. InCluster 配置
// 3. 本地 KUBECONFIG 或 ~/.kube/config
// 返回值第二个布尔含义：是否存在可用集群（用于决定是否进行选举）
func getClientset(clusterID string) (*kubernetes.Clientset, bool, error) {
	// 如果提供了 ClusterID，优先尝试使用指定集群的配置
	if clusterID != "" {
		klog.V(6).Infof("[leader] 尝试使用指定的 ClusterID 获取配置：%s", clusterID)
		// 通过服务获取集群配置
		cluster := service.ClusterService().GetClusterByID(clusterID)
		if cluster == nil {
			klog.V(6).Infof("[leader] 未找到指定的集群配置：%s，继续尝试 InCluster 配置", clusterID)
		} else {
			restCfg := cluster.GetRestConfig()
			if restCfg != nil {
				clientset, err := kubernetes.NewForConfig(restCfg)
				if err == nil {
					klog.V(6).Infof("[leader] 已使用指定集群的配置创建 ClientSet：%s", clusterID)
					// 找到可用集群
					return clientset, true, nil
				}
				klog.V(6).Infof("[leader] 使用指定集群创建 ClientSet 失败：%v，继续尝试 InCluster 配置", err)
			}
			klog.V(6).Infof("[leader] 指定集群未提供 RestConfig：%s，继续尝试 InCluster 配置", clusterID)
		}
	}
	// 尝试 InCluster 模式
	config, err := rest.InClusterConfig()
	if err == nil {
		clientset, cErr := kubernetes.NewForConfig(config)
		if cErr == nil {
			return clientset, true, nil
		}
		klog.V(6).Infof("[leader] InCluster 创建 ClientSet 失败：%v，继续尝试本地 kubeconfig", cErr)
	}

	// Fallback 到本地 kubeconfig
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		kubeconfig = os.ExpandEnv("$HOME/.kube/config")
	}
	config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		klog.V(6).Infof("[leader] 本地 kubeconfig 加载失败：%v", err)
		// 无可用集群，返回非错误以便降级为本地 Leader
		return nil, false, nil
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		klog.V(6).Infof("[leader] 本地 kubeconfig 创建 ClientSet 失败：%v", err)
		return nil, false, nil
	}
	return clientset, true, nil
}

// detectNamespace 尝试自动检测当前运行的 Namespace
// 在 InCluster 模式下从 ServiceAccount 文件读取，否则默认 "default"。
func detectNamespace() string {
	data, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err == nil {
		return string(data)
	}
	return "default"
}
