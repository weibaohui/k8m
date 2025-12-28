package webhook

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"k8s.io/klog/v2"
)

type WebhookLifecycle struct{}

func (w *WebhookLifecycle) Install(ctx plugins.InstallContext) error {
	klog.V(6).Infof("安装Webhook插件成功")
	return nil
}

func (w *WebhookLifecycle) Upgrade(ctx plugins.UpgradeContext) error {
	klog.V(6).Infof("升级Webhook插件：从版本 %s 到版本 %s", ctx.FromVersion(), ctx.ToVersion())
	return nil
}

func (w *WebhookLifecycle) Enable(ctx plugins.EnableContext) error {
	klog.V(6).Infof("启用Webhook插件")
	return nil
}

func (w *WebhookLifecycle) Disable(ctx plugins.BaseContext) error {
	klog.V(6).Infof("禁用Webhook插件")
	return nil
}

func (w *WebhookLifecycle) Uninstall(ctx plugins.InstallContext) error {
	klog.V(6).Infof("卸载Webhook插件成功")
	return nil
}

func (w *WebhookLifecycle) Start(ctx plugins.BaseContext) error {
	return nil
}

func (w *WebhookLifecycle) StartCron(ctx plugins.BaseContext, spec string) error {
	return nil
}
