package webhook

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules/webhook/models"
	"k8s.io/klog/v2"
)

type WebhookLifecycle struct{}

func (w *WebhookLifecycle) Install(ctx plugins.InstallContext) error {
	klog.V(6).Infof("开始安装Webhook插件")
	if err := models.InitDB(); err != nil {
		klog.V(6).Infof("安装Webhook插件失败，初始化数据库失败: %v", err)
		return err
	}
	klog.V(6).Infof("安装Webhook插件成功")
	return nil
}

func (w *WebhookLifecycle) Upgrade(ctx plugins.UpgradeContext) error {
	klog.V(6).Infof("升级Webhook插件：从版本 %s 到版本 %s", ctx.FromVersion(), ctx.ToVersion())
	if err := models.UpgradeDB(ctx.FromVersion(), ctx.ToVersion()); err != nil {
		klog.V(6).Infof("升级Webhook插件失败: %v", err)
		return err
	}
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

func (w *WebhookLifecycle) Uninstall(ctx plugins.UninstallContext) error {
	klog.V(6).Infof("开始卸载Webhook插件")

	// 根据keepData参数决定是否删除数据库
	if !ctx.KeepData() {
		if err := models.DropDB(); err != nil {
			klog.V(6).Infof("卸载事件转发插件失败: %v", err)
			return err
		}
	}
	klog.V(6).Infof("卸载Webhook插件成功")
	return nil
}

func (w *WebhookLifecycle) Start(ctx plugins.BaseContext) error {
	RegisterAllAdapters()
	klog.V(6).Infof("启动Webhook插件成功")
	return nil
}

func (w *WebhookLifecycle) StartCron(ctx plugins.BaseContext, spec string) error {
	return nil
}
