package leader

import (
    "context"
    "fmt"
    "os"
    "time"

    "github.com/weibaohui/k8m/pkg/comm/utils"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
    clientset, hasCluster, err := utils.GetClientSet(cfg.ClusterID)
    if err != nil {
        return fmt.Errorf("get clientset failed: %w", err)
    }

    if cfg.Namespace == "" {
        cfg.Namespace = utils.DetectNamespace()
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
