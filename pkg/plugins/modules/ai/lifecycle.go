package ai

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"k8s.io/klog/v2"
)

type AILifecycle struct{}

func (l *AILifecycle) Install(ctx plugins.InstallContext) error {
	klog.V(6).Infof("安装 AI 插件成功")
	return nil
}

func (l *AILifecycle) Upgrade(ctx plugins.UpgradeContext) error {
	klog.V(6).Infof("升级 AI 插件：从版本 %s 到版本 %s", ctx.FromVersion(), ctx.ToVersion())
	return nil
}

func (l *AILifecycle) Enable(ctx plugins.EnableContext) error {
	klog.V(6).Infof("启用 AI 插件")
	return nil
}

func (l *AILifecycle) Disable(ctx plugins.BaseContext) error {
	klog.V(6).Infof("禁用 AI 插件")
	return nil
}

func (l *AILifecycle) Uninstall(ctx plugins.UninstallContext) error {
	klog.V(6).Infof("卸载 AI 插件成功")
	return nil
}

func (l *AILifecycle) Start(ctx plugins.BaseContext) error {
	klog.V(6).Infof("启动 AI 插件后台任务")
	return nil
}

func (l *AILifecycle) StartCron(ctx plugins.BaseContext, spec string) error {
	return nil
}
