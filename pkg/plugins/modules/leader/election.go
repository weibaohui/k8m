package leader

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/weibaohui/k8m/pkg/comm/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
	"k8s.io/klog/v2"
)

// Config Leader选举配置
// 包含命名空间、锁名称、选举时长、续租截止、重试周期以及领导开始/结束回调。
// 要求：LeaseDuration > RenewDeadline > RetryPeriod * 3
type Config struct {
	Namespace        string
	LockName         string
	LeaseDuration    time.Duration // 租约持续时间
	RenewDeadline    time.Duration // 续租截止时间
	RetryPeriod      time.Duration // 重试周期
	ClusterID        string        // 指定用于选举的宿主集群ID，留空时自动检测
	OnStartedLeading func(ctx context.Context)
	OnStoppedLeading func()
}

// Run 启动Leader选举逻辑
// 选择优先级：指定ClusterID -> InCluster -> 本地kubeconfig；均不可用时退化为本地Leader（不进行选举）
func Run(ctx context.Context, cfg Config) error {
	clientset, hasCluster, err := utils.GetClientSet(cfg.ClusterID)
	if err != nil {
		return fmt.Errorf("获取ClientSet失败: %w", err)
	}

	if cfg.Namespace == "" {
		cfg.Namespace = utils.DetectNamespace()
	}

	// 非集群模式：没有指定集群、不是InCluster、也没有本地kubeconfig
	if !hasCluster {
		klog.V(6).Infof("无可用的K8s集群，直接作为Leader运行（不进行选举）")
		_ = patchPodRoleLabel(ctx, clientset, "leader")
		if cfg.OnStartedLeading != nil {
			cfg.OnStartedLeading(ctx)
		}
		return nil
	}

	// 默认参数
	if cfg.LeaseDuration == 0 {
		cfg.LeaseDuration = 15 * time.Second
	}
	if cfg.RenewDeadline == 0 {
		cfg.RenewDeadline = 10 * time.Second
	}
	if cfg.RetryPeriod == 0 {
		cfg.RetryPeriod = 2 * time.Second
	}

	id := utils.GenerateInstanceID()
	klog.V(6).Infof("当前实例选举ID: %s", id)
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
			// 成为Leader
			OnStartedLeading: func(c context.Context) {
				podName := os.Getenv("POD_NAME")
				namespace := os.Getenv("POD_NAMESPACE")
				podIP := os.Getenv("POD_IP")
				klog.V(6).Infof("开始作为Leader运行,Leader身份: %s/%s(%s)", namespace, podName, podIP)
				if err := patchPodRoleLabel(c, clientset, "leader"); err != nil {
					klog.Errorf("设置 Pod %s/%s(%s) leader label 失败: %v", namespace, podName, podIP, err)
				}
				if cfg.OnStartedLeading != nil {
					cfg.OnStartedLeading(c)
				}
			},
			// 失去Leader
			OnStoppedLeading: func() {
				klog.V(6).Infof("停止作为Leader运行")
				pctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
				defer cancel()
				if err := patchPodRoleLabel(pctx, clientset, "follower"); err != nil {
					klog.Errorf("恢复 follower label 失败: %v", err)
				}
				if cfg.OnStoppedLeading != nil {
					cfg.OnStoppedLeading()
				}
			},
			// 新Leader产生
			OnNewLeader: func(identity string) {
				if identity == id {
					klog.V(6).Infof("我成为新的Leader：%s", id)
				} else {
					klog.V(6).Infof("选举产生新的Leader：%s", identity)
				}
			},
		},
	}

	klog.V(6).Infof("开始进行Leader选举（锁=%s/%s）", cfg.Namespace, cfg.LockName)
	leaderelection.RunOrDie(ctx, leaderelectionCfg)
	return nil
}
func patchPodRoleLabel(
	ctx context.Context,
	clientset kubernetes.Interface,
	role string,
) error {
	podName := os.Getenv("POD_NAME")
	namespace := os.Getenv("POD_NAMESPACE")

	if podName == "" || namespace == "" {
		return fmt.Errorf("POD_NAME or POD_NAMESPACE not set")
	}

	patch := fmt.Appendf(nil, `{
	  "metadata": {
	    "labels": {
	      "k8m.io/role": "%s"
	    }
	  }
	}`, role)

	_, err := clientset.CoreV1().
		Pods(namespace).
		Patch(
			ctx,
			podName,
			types.MergePatchType,
			patch,
			metav1.PatchOptions{},
		)

	if err != nil {
		return fmt.Errorf("patch pod label failed: %w", err)
	}

	klog.V(6).Infof(
		"patched pod %s/%s label k8m.io/role=%s",
		namespace, podName, role,
	)

	return nil
}
