package k8swatch

import (
	"context"

	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/eventbus"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/service"
	"k8s.io/klog/v2"
)

type K8sWatchLifecycle struct {
	leaderWatchCancel context.CancelFunc
}

func (k *K8sWatchLifecycle) Install(ctx plugins.InstallContext) error {
	klog.V(6).Infof("安装K8sWatch插件成功")
	return nil
}

func (k *K8sWatchLifecycle) Upgrade(ctx plugins.UpgradeContext) error {
	klog.V(6).Infof("升级K8sWatch插件：从版本 %s 到版本 %s", ctx.FromVersion(), ctx.ToVersion())
	return nil
}

func (k *K8sWatchLifecycle) Enable(ctx plugins.EnableContext) error {
	klog.V(6).Infof("启用K8sWatch插件")
	return nil
}

func (k *K8sWatchLifecycle) Disable(ctx plugins.BaseContext) error {
	klog.V(6).Infof("禁用K8sWatch插件")
	return nil
}

func (k *K8sWatchLifecycle) Uninstall(ctx plugins.UninstallContext) error {
	klog.V(6).Infof("卸载K8sWatch插件")
	return nil
}

func (k *K8sWatchLifecycle) Start(ctx plugins.BaseContext) error {
	if plugins.ManagerInstance().IsRunning(modules.PluginNameLeader) {
		elect := ctx.Bus().Subscribe(eventbus.EventLeaderElected)
		lost := ctx.Bus().Subscribe(eventbus.EventLeaderLost)

		leaderWatchCtx, cancel := context.WithCancel(context.Background())
		k.leaderWatchCancel = cancel

		go func() {
			for {
				select {
				case <-elect:
					klog.V(6).Infof("成为Leader，启动K8s资源监听")
					k.startWatch()
				case <-lost:
					klog.V(6).Infof("不再是Leader，停止K8s资源监听")
					k.stopWatch()
				case <-leaderWatchCtx.Done():
					klog.V(6).Infof("K8sWatch插件 Leader 监听 goroutine 退出")
					return
				}
			}
		}()

		klog.V(6).Infof("根据实例Leader状态启动K8sWatch插件后台任务")
	} else {
		k.startWatch()
		klog.V(6).Infof("启动K8sWatch插件后台任务")
	}
	return nil
}

func (k *K8sWatchLifecycle) StartCron(ctx plugins.BaseContext, spec string) error {
	klog.V(6).Infof("执行K8sWatch插件定时任务，表达式: %s", spec)
	return nil
}

func (k *K8sWatchLifecycle) Stop(ctx plugins.BaseContext) error {
	klog.V(6).Infof("停止K8sWatch插件后台任务")

	if k.leaderWatchCancel != nil {
		k.leaderWatchCancel()
		k.leaderWatchCancel = nil
	}

	k.stopWatch()
	return nil
}

func (k *K8sWatchLifecycle) startWatch() {
	service.ClusterService().DelayStartFunc(func() {
		service.PodService().Watch()
		service.NodeService().Watch()
		service.PVCService().Watch()
		service.PVService().Watch()
		service.IngressService().Watch()
	})
}

func (k *K8sWatchLifecycle) stopWatch() {
	klog.V(6).Infof("停止K8s资源监听")
}
