package eventhandler

import (
	"context"

	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/eventbus"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/eventhandler/models"
	"k8s.io/klog/v2"
)

// EventHandlerLifecycle 中文函数注释：事件转发插件生命周期实现。
type EventHandlerLifecycle struct {
	leaderWatchCancel context.CancelFunc
}

// Install 中文函数注释：安装事件转发插件，初始化数据库表结构。
func (l *EventHandlerLifecycle) Install(ctx plugins.InstallContext) error {
	if err := models.InitDB(); err != nil {
		klog.V(6).Infof("安装事件转发插件失败: %v", err)
		return err
	}
	klog.V(6).Infof("安装事件转发插件成功")
	return nil
}

// Upgrade 中文函数注释：升级事件转发插件，执行必要的数据库迁移。
func (l *EventHandlerLifecycle) Upgrade(ctx plugins.UpgradeContext) error {
	klog.V(6).Infof("升级事件转发插件：从版本 %s 到版本 %s", ctx.FromVersion(), ctx.ToVersion())
	if err := models.UpgradeDB(ctx.FromVersion(), ctx.ToVersion()); err != nil {
		klog.V(6).Infof("升级事件转发插件失败: %v", err)
		return err
	}
	return nil
}

// Enable 中文函数注释：启用事件转发插件，确保数据库表存在。
func (l *EventHandlerLifecycle) Enable(ctx plugins.EnableContext) error {
	if err := models.InitDB(); err != nil {
		klog.V(6).Infof("启用事件转发插件失败: %v", err)
		return err
	}
	klog.V(6).Infof("启用事件转发插件")
	return nil
}

// Disable 中文函数注释：禁用事件转发插件，停止后台任务与事件转发。
func (l *EventHandlerLifecycle) Disable(ctx plugins.BaseContext) error {
	klog.V(6).Infof("禁用事件转发插件")
	return nil
}

// Uninstall 中文函数注释：卸载事件转发插件，停止后台任务并根据keepData参数决定是否删除相关表。
func (l *EventHandlerLifecycle) Uninstall(ctx plugins.UninstallContext) error {
	// 根据keepData参数决定是否删除数据库
	if !ctx.KeepData() {
		if err := models.DropDB(); err != nil {
			klog.V(6).Infof("卸载事件转发插件失败: %v", err)
			return err
		}
	}
	klog.V(6).Infof("卸载事件转发插件成功")
	return nil
}

// Start 中文函数注释：启动事件转发插件后台任务（不可阻塞），按主备状态控制事件转发启停。
func (l *EventHandlerLifecycle) Start(ctx plugins.BaseContext) error {
	if plugins.ManagerInstance().IsEnabled(modules.PluginNameLeader) {
		elect := ctx.Bus().Subscribe(eventbus.EventLeaderElected)
		lost := ctx.Bus().Subscribe(eventbus.EventLeaderLost)

		leaderWatchCtx, cancel := context.WithCancel(context.Background())
		l.leaderWatchCancel = cancel

		go func() {
			for {
				select {
				case <-elect:
					StartLeaderWatch()
					klog.V(6).Infof("成为Leader，启动事件转发")
				case <-lost:
					StopLeaderWatch()
					klog.V(6).Infof("不再是Leader，停止事件转发")
				case <-leaderWatchCtx.Done():
					klog.V(6).Infof("事件转发插件 Leader 监听 goroutine 退出")
					return
				}
			}
		}()
		klog.V(6).Infof("根据实例Leader获取状态启动事件转发插件后台任务")
	} else {
		StartEventForwardingWatch()
		klog.V(6).Infof("启动事件转发插件后台任务")
	}
	return nil
}

// StartCron 中文函数注释：事件转发插件不使用插件级 cron，留空实现。
func (l *EventHandlerLifecycle) StartCron(ctx plugins.BaseContext, spec string) error {
	return nil
}

// Stop 停止事件转发插件的后台任务
func (l *EventHandlerLifecycle) Stop(ctx plugins.BaseContext) error {
	klog.V(6).Infof("停止事件转发插件后台任务")

	if l.leaderWatchCancel != nil {
		l.leaderWatchCancel()
		l.leaderWatchCancel = nil
	}

	StopLeaderWatch()
	return nil
}
