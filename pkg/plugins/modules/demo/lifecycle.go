package demo

import (
	"context"
	"time"

	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules/demo/models"
	"k8s.io/klog/v2"
)

// DemoLifecycle Demo插件生命周期实现
type DemoLifecycle struct {
	cancelStart context.CancelFunc
}

// Install 安装Demo插件，初始化数据库表
func (d *DemoLifecycle) Install(ctx plugins.InstallContext) error {
	if err := models.InitDB(); err != nil {
		klog.V(6).Infof("安装Demo插件失败: %v", err)
		return err
	}
	klog.V(6).Infof("安装Demo插件成功")
	return nil
}

// Upgrade 升级Demo插件，执行数据库迁移
func (d *DemoLifecycle) Upgrade(ctx plugins.UpgradeContext) error {
	klog.V(6).Infof("升级Demo插件：从版本 %s 到版本 %s", ctx.FromVersion(), ctx.ToVersion())
	if err := models.UpgradeDB(ctx.FromVersion(), ctx.ToVersion()); err != nil {
		return err
	}
	return nil
}

// Enable 启用Demo插件
func (d *DemoLifecycle) Enable(ctx plugins.EnableContext) error {
	klog.V(6).Infof("启用Demo插件")
	return nil
}

// Disable 禁用Demo插件
func (d *DemoLifecycle) Disable(ctx plugins.BaseContext) error {
	klog.V(6).Infof("禁用Demo插件")

	if d.cancelStart != nil {
		d.cancelStart()
		d.cancelStart = nil
	}

	return nil
}

// Uninstall 卸载Demo插件，根据keepData参数决定是否删除相关的表及数据
func (d *DemoLifecycle) Uninstall(ctx plugins.UninstallContext) error {
	klog.V(6).Infof("卸载Demo插件")
	// 根据keepData参数决定是否删除数据库
	if !ctx.KeepData() {
		if err := models.DropDB(); err != nil {
			return err
		}
		klog.V(6).Infof("卸载Demo插件完成，已删除相关表及数据")
	} else {
		klog.V(6).Infof("卸载Demo插件完成，保留相关表及数据")
	}
	return nil
}

// Start 启动Demo插件的后台任务（不可阻塞）
// 该方法由系统在 Manager.Start 时统一调用，用于启动非阻塞的后台协程或定时任务
func (d *DemoLifecycle) Start(ctx plugins.BaseContext) error {
	klog.V(6).Infof("启动Demo插件后台任务")

	startCtx, cancel := context.WithCancel(context.Background())
	d.cancelStart = cancel

	go func(meta plugins.Meta) {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				klog.V(6).Infof("Demo插件后台任务运行中，插件: %s，版本: %s", meta.Name, meta.Version)
			case <-startCtx.Done():
				klog.V(6).Infof("Demo 插件启动 goroutine 退出")
				return
			}
		}
	}(ctx.Meta())

	return nil
}

// StartCron 启动Demo插件的定时任务（不可阻塞）
// 该方法由系统根据 metadata 中定义的 5 段 cron 表达式触发
func (d *DemoLifecycle) StartCron(ctx plugins.BaseContext, spec string) error {
	klog.V(6).Infof("执行Demo插件定时任务，表达式: %s", spec)
	return nil
}

// Stop 停止Demo插件的后台任务
func (d *DemoLifecycle) Stop(ctx plugins.BaseContext) error {
	klog.V(6).Infof("停止Demo插件后台任务")

	if d.cancelStart != nil {
		d.cancelStart()
		d.cancelStart = nil
	}

	return nil
}
