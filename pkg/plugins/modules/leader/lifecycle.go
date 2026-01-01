package leader

import (
	"context"
	"time"

	helm2 "github.com/weibaohui/k8m/pkg/helm"
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/eventbus"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/inspection/lua"
	"github.com/weibaohui/k8m/pkg/service"
	"k8s.io/klog/v2"
)

// LeaderLifecycle Leader选举插件生命周期实现
// 负责启动Leader选举，以及在成为Leader时启动/停止平台的后台任务
type LeaderLifecycle struct {
	// cleanupCancel 用于控制Leader清理任务的停止
	cleanupCancel context.CancelFunc
}

// Install 安装Leader选举插件
// 该插件不涉及数据库初始化，安装阶段仅打印日志
func (l *LeaderLifecycle) Install(ctx plugins.InstallContext) error {
	klog.V(6).Infof("安装Leader选举插件")
	return nil
}

// Upgrade 升级Leader选举插件
// 该插件暂无升级数据库的需求，升级阶段仅打印日志
func (l *LeaderLifecycle) Upgrade(ctx plugins.UpgradeContext) error {
	klog.V(6).Infof("升级Leader选举插件：从版本 %s 到版本 %s", ctx.FromVersion(), ctx.ToVersion())
	return nil
}

// Enable 启用Leader选举插件
// 启用阶段仅打印日志，真正的后台任务由 Start 中的选举逻辑管理
func (l *LeaderLifecycle) Enable(ctx plugins.EnableContext) error {
	klog.V(6).Infof("启用Leader选举插件")
	return nil
}

// Disable 禁用Leader选举插件
// 禁用阶段仅打印日志；选举停止与任务收敛由选举停止回调处理
func (l *LeaderLifecycle) Disable(ctx plugins.BaseContext) error {
	klog.V(6).Infof("禁用Leader选举插件")
	return nil
}

// Uninstall 卸载Leader选举插件
// 该插件不涉及可删除的持久化资源，卸载阶段仅打印日志
func (l *LeaderLifecycle) Uninstall(ctx plugins.UninstallContext) error {
	klog.V(6).Infof("卸载Leader选举插件")
	return nil
}

// Start 启动Leader选举的后台任务（不可阻塞）
// 由插件管理器在系统启动时统一调用，用于启动选举并在成为Leader时执行平台任务
func (l *LeaderLifecycle) Start(ctx plugins.BaseContext) error {
	klog.V(6).Infof("启动Leader选举插件后台任务")

	go func() {
		leaderCfg := Config{
			LockName:      "k8m-leader-lock",
			LeaseDuration: 60 * time.Second,
			RenewDeadline: 50 * time.Second,
			RetryPeriod:   10 * time.Second,
			OnStartedLeading: func(c context.Context) {
				klog.V(6).Infof("成为Leader，启动定时任务（集群巡检、Helm仓库更新）")
				service.LeaderService().SetCurrentLeader(true)
				ctx.Bus().Publish(eventbus.Event{
					Type: eventbus.EventLeaderElected,
				})
				// 只有当集群巡检插件启用时才启动巡检定时任务
				if plugins.ManagerInstance().IsEnabled(modules.PluginNameInspection) {
					lua.InitClusterInspection()
				}
				helm2.StartUpdateHelmRepoInBackground()
			},
			OnStoppedLeading: func() {
				klog.V(6).Infof("不再是Leader，停止定时任务（集群巡检、Helm仓库更新）")
				service.LeaderService().SetCurrentLeader(false)
				ctx.Bus().Publish(eventbus.Event{
					Type: eventbus.EventLeaderLost,
				})
				// 只有当集群巡检插件启用时才停止巡检定时任务
				if plugins.ManagerInstance().IsEnabled(modules.PluginNameInspection) {
					lua.StopClusterInspection()
				}
				helm2.StopUpdateHelmRepoInBackground()
			},
		}
		bg := context.Background()
		if err := Run(bg, leaderCfg); err != nil {
			klog.V(6).Infof("Leader选举失败: %v", err)
		}
	}()
	return nil
}

// StartCron 该插件不使用定时任务，留空实现
func (l *LeaderLifecycle) StartCron(ctx plugins.BaseContext, spec string) error {
	return nil
}
