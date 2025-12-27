package demo

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules/demo/models"
	"k8s.io/klog/v2"
	"time"
)

// DemoLifecycle Demo插件生命周期实现
type DemoLifecycle struct{}

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
	return nil
}

// Uninstall 卸载Demo插件，删除相关的表及数据
func (d *DemoLifecycle) Uninstall(ctx plugins.InstallContext) error {
	klog.V(6).Infof("卸载Demo插件，删除相关的表及数据")
	if err := models.DropDB(); err != nil {
		return err
	}
	klog.V(6).Infof("卸载Demo插件完成，已删除相关表及数据")
	return nil
}

// Start 启动Demo插件的后台任务（不可阻塞）
// 该方法由系统在 Manager.Start 时统一调用，用于启动非阻塞的后台协程或定时任务
func (d *DemoLifecycle) Start(ctx plugins.BaseContext) error {
	klog.V(6).Infof("启动Demo插件后台任务")
	// 启动一个简单的定时任务示例：定期输出运行日志
	go func(meta plugins.Meta) {
		ticker := time.NewTicker(30 * time.Second)
		for range ticker.C {
			klog.V(6).Infof("Demo插件后台任务运行中，插件: %s，版本: %s", meta.Name, meta.Version)
		}
	}(ctx.Meta())
	return nil
}
