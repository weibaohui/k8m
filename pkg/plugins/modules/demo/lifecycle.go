package demo

import (
	"github.com/weibaohui/k8m/pkg/plugins/modules/demo/backend"
	"github.com/weibaohui/k8m/pkg/plugins"
	"k8s.io/klog/v2"
)

// DemoLifecycle Demo插件生命周期实现
type DemoLifecycle struct{}

// Install 安装Demo插件，初始化数据库表
func (d *DemoLifecycle) Install(ctx plugins.InstallContext) error {
	if err := backend.InitDB(); err != nil {
		klog.V(6).Infof("安装Demo插件失败: %v", err)
		return err
	}
	klog.V(6).Infof("安装Demo插件成功")
	return nil
}

// Upgrade 升级Demo插件，当前未实现
func (d *DemoLifecycle) Upgrade(ctx plugins.UpgradeContext) error {
	klog.V(6).Infof("升级Demo插件：从版本 %s 到版本 %s", ctx.FromVersion(), ctx.ToVersion())
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

// Uninstall 卸载Demo插件，当前不做删除数据
func (d *DemoLifecycle) Uninstall(ctx plugins.InstallContext) error {
	klog.V(6).Infof("卸载Demo插件")
	return nil
}
