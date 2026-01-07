package helm

import (
	"context"
	"sync"

	"github.com/robfig/cron/v3"
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/eventbus"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/helm/models"
	helm "github.com/weibaohui/k8m/pkg/plugins/modules/helm/service"
	"k8s.io/klog/v2"
)

// HelmLifecycle Helm 插件生命周期实现
type HelmLifecycle struct {
	helmCron          *cron.Cron
	helmMu            sync.Mutex
	leaderWatchMu     sync.Mutex
	leaderWatchCancel context.CancelFunc
}

// Install 安装 Helm 插件
func (l *HelmLifecycle) Install(ctx plugins.InstallContext) error {
	if err := models.InitDB(); err != nil {
		klog.V(6).Infof("安装 Helm 插件失败: %v", err)
		return err
	}
	klog.V(6).Infof("安装 Helm 插件成功")
	return nil
}

// Upgrade 升级 Helm 插件
func (l *HelmLifecycle) Upgrade(ctx plugins.UpgradeContext) error {
	klog.V(6).Infof("升级 Helm 插件：从版本 %s 到版本 %s", ctx.FromVersion(), ctx.ToVersion())
	if err := models.UpgradeDB(ctx.FromVersion(), ctx.ToVersion()); err != nil {
		klog.V(6).Infof("升级 Helm 插件失败: %v", err)
		return err
	}
	return nil
}

// Enable 启用 Helm 插件
func (l *HelmLifecycle) Enable(ctx plugins.EnableContext) error {
	klog.V(6).Infof("启用 Helm 插件")
	return nil
}

// Disable 禁用 Helm 插件
func (l *HelmLifecycle) Disable(ctx plugins.BaseContext) error {
	klog.V(6).Infof("禁用 Helm 插件")
	return nil
}

// Uninstall 卸载 Helm 插件
func (l *HelmLifecycle) Uninstall(ctx plugins.UninstallContext) error {
	// 根据keepData参数决定是否删除数据库
	if !ctx.KeepData() {
		if err := models.DropDB(); err != nil {
			klog.V(6).Infof("卸载 Helm 插件失败: %v", err)
			return err
		}
	}
	klog.V(6).Infof("卸载 Helm 插件成功")
	return nil
}

// Start 启动 Helm 插件后台任务（不可阻塞）
func (l *HelmLifecycle) Start(ctx plugins.BaseContext) error {
	if plugins.ManagerInstance().IsRunning(modules.PluginNameLeader) {
		// 如果启用了 Leader 插件，监听 Leader 选举事件
		elect := ctx.Bus().Subscribe(eventbus.EventLeaderElected)
		lost := ctx.Bus().Subscribe(eventbus.EventLeaderLost)

		leaderWatchCtx, cancel := context.WithCancel(context.Background())
		l.leaderWatchCancel = cancel

		go func() {
			for {
				select {
				case <-elect:
					klog.V(6).Infof("成为Leader，启动 Helm 仓库更新定时任务")
					helm.StartUpdateHelmRepoInBackground()
				case <-lost:
					klog.V(6).Infof("不再是Leader，停止 Helm 仓库更新定时任务")
					helm.StopUpdateHelmRepoInBackground()
				case <-leaderWatchCtx.Done():
					klog.V(6).Infof("Helm 插件 Leader 监听 goroutine 退出")
					return
				}
			}
		}()

		klog.V(6).Infof("根据实例Leader状态启动 Helm 插件后台任务")
	} else {
		// 没有启用 Leader 插件，直接启动定时任务
		helm.StartUpdateHelmRepoInBackground()
		klog.V(6).Infof("启动 Helm 插件后台任务")
	}
	return nil
}

// StartCron Helm 插件使用自定义定时任务，留空实现
func (l *HelmLifecycle) StartCron(ctx plugins.BaseContext, spec string) error {
	return nil
}

// Stop 停止 Helm 插件的后台任务
func (l *HelmLifecycle) Stop(ctx plugins.BaseContext) error {
	klog.V(6).Infof("停止 Helm 插件后台任务")

	if l.leaderWatchCancel != nil {
		l.leaderWatchCancel()
		l.leaderWatchCancel = nil
	}

	helm.StopUpdateHelmRepoInBackground()
	return nil
}
