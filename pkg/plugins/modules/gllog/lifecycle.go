package gllog

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"k8s.io/klog/v2"
)

type GlobalLogLifecycle struct{}

func (g *GlobalLogLifecycle) Install(ctx plugins.InstallContext) error {
	klog.V(6).Infof("安装全局日志插件")
	return nil
}

func (g *GlobalLogLifecycle) Upgrade(ctx plugins.UpgradeContext) error {
	klog.V(6).Infof("升级全局日志插件：从版本 %s 到版本 %s", ctx.FromVersion(), ctx.ToVersion())
	return nil
}

func (g *GlobalLogLifecycle) Enable(ctx plugins.EnableContext) error {
	klog.V(6).Infof("启用全局日志插件")
	return nil
}

func (g *GlobalLogLifecycle) Disable(ctx plugins.BaseContext) error {
	klog.V(6).Infof("禁用全局日志插件")
	return nil
}

func (g *GlobalLogLifecycle) Uninstall(ctx plugins.UninstallContext) error {
	klog.V(6).Infof("卸载全局日志插件")
	return nil
}

func (g *GlobalLogLifecycle) Start(ctx plugins.BaseContext) error {
	klog.V(6).Infof("启动全局日志插件")
	return nil
}

func (g *GlobalLogLifecycle) StartCron(ctx plugins.BaseContext, spec string) error {
	klog.V(6).Infof("执行全局日志插件定时任务，表达式: %s", spec)
	return nil
}
