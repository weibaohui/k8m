package kubeconfig_export

import (
	"context"

	"github.com/weibaohui/k8m/pkg/plugins"
	"k8s.io/klog/v2"
)

// KubeconfigExportLifecycle Kubeconfig导出插件生命周期实现
type KubeconfigExportLifecycle struct {
	cancelStart context.CancelFunc
}

// Install 安装Kubeconfig导出插件
func (k *KubeconfigExportLifecycle) Install(ctx plugins.InstallContext) error {
	klog.V(6).Infof("安装Kubeconfig导出插件成功")
	return nil
}

// Upgrade 升级Kubeconfig导出插件
func (k *KubeconfigExportLifecycle) Upgrade(ctx plugins.UpgradeContext) error {
	klog.V(6).Infof("升级Kubeconfig导出插件：从版本 %s 到版本 %s", ctx.FromVersion(), ctx.ToVersion())
	return nil
}

// Enable 启用Kubeconfig导出插件
func (k *KubeconfigExportLifecycle) Enable(ctx plugins.EnableContext) error {
	klog.V(6).Infof("启用Kubeconfig导出插件")
	return nil
}

// Disable 禁用Kubeconfig导出插件
func (k *KubeconfigExportLifecycle) Disable(ctx plugins.BaseContext) error {
	klog.V(6).Infof("禁用Kubeconfig导出插件")
	return nil
}

// Uninstall 卸载Kubeconfig导出插件
func (k *KubeconfigExportLifecycle) Uninstall(ctx plugins.UninstallContext) error {
	klog.V(6).Infof("卸载Kubeconfig导出插件完成")
	return nil
}

// Start 启动Kubeconfig导出插件的后台任务
func (k *KubeconfigExportLifecycle) Start(ctx plugins.BaseContext) error {
	klog.V(6).Infof("启动Kubeconfig导出插件后台任务")
	return nil
}

// StartCron 执行Kubeconfig导出插件的定时任务
func (k *KubeconfigExportLifecycle) StartCron(ctx plugins.BaseContext, spec string) error {
	klog.V(6).Infof("执行Kubeconfig导出插件定时任务，表达式: %s", spec)
	return nil
}

// Stop 停止Kubeconfig导出插件的后台任务
func (k *KubeconfigExportLifecycle) Stop(ctx plugins.BaseContext) error {
	klog.V(6).Infof("停止Kubeconfig导出插件后台任务")
	return nil
}