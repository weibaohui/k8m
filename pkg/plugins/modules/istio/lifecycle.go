package istio

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"k8s.io/klog/v2"
)

type IstioLifecycle struct {
}

func (i *IstioLifecycle) Install(ctx plugins.InstallContext) error {
	klog.V(6).Infof("安装Istio插件成功")
	return nil
}

func (i *IstioLifecycle) Upgrade(ctx plugins.UpgradeContext) error {
	klog.V(6).Infof("升级Istio插件：从版本 %s 到版本 %s", ctx.FromVersion(), ctx.ToVersion())
	return nil
}

func (i *IstioLifecycle) Enable(ctx plugins.EnableContext) error {
	klog.V(6).Infof("启用Istio插件")
	return nil
}

func (i *IstioLifecycle) Disable(ctx plugins.BaseContext) error {
	klog.V(6).Infof("禁用Istio插件")
	return nil
}

func (i *IstioLifecycle) Uninstall(ctx plugins.UninstallContext) error {
	klog.V(6).Infof("卸载Istio插件")
	if !ctx.KeepData() {
		klog.V(6).Infof("卸载Istio插件完成，已删除相关表及数据")
	} else {
		klog.V(6).Infof("卸载Istio插件完成，保留相关表及数据")
	}
	return nil
}

func (i *IstioLifecycle) Start(ctx plugins.BaseContext) error {
	klog.V(6).Infof("启动Istio插件后台任务")
	return nil
}

func (i *IstioLifecycle) StartCron(ctx plugins.BaseContext, spec string) error {
	klog.V(6).Infof("执行Istio插件定时任务，表达式: %s", spec)
	return nil
}

func (i *IstioLifecycle) Stop(ctx plugins.BaseContext) error {
	klog.V(6).Infof("停止Istio插件后台任务")
	return nil
}
