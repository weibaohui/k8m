package openkruise

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"k8s.io/klog/v2"
)

type OpenKruiseLifecycle struct {
}

func (o *OpenKruiseLifecycle) Install(ctx plugins.InstallContext) error {
	klog.V(6).Infof("安装OpenKruise插件成功")
	return nil
}

func (o *OpenKruiseLifecycle) Upgrade(ctx plugins.UpgradeContext) error {
	klog.V(6).Infof("升级OpenKruise插件：从版本 %s 到版本 %s", ctx.FromVersion(), ctx.ToVersion())
	return nil
}

func (o *OpenKruiseLifecycle) Enable(ctx plugins.EnableContext) error {
	klog.V(6).Infof("启用OpenKruise插件")
	return nil
}

func (o *OpenKruiseLifecycle) Disable(ctx plugins.BaseContext) error {
	klog.V(6).Infof("禁用OpenKruise插件")
	return nil
}

func (o *OpenKruiseLifecycle) Uninstall(ctx plugins.UninstallContext) error {
	klog.V(6).Infof("卸载OpenKruise插件")
	if !ctx.KeepData() {
		klog.V(6).Infof("卸载OpenKruise插件完成，已删除相关表及数据")
	} else {
		klog.V(6).Infof("卸载OpenKruise插件完成，保留相关表及数据")
	}
	return nil
}

func (o *OpenKruiseLifecycle) Start(ctx plugins.BaseContext) error {
	klog.V(6).Infof("启动OpenKruise插件后台任务")
	return nil
}

func (o *OpenKruiseLifecycle) StartCron(ctx plugins.BaseContext, spec string) error {
	klog.V(6).Infof("执行OpenKruise插件定时任务，表达式: %s", spec)
	return nil
}

func (o *OpenKruiseLifecycle) Stop(ctx plugins.BaseContext) error {
	klog.V(6).Infof("停止OpenKruise插件后台任务")
	return nil
}
