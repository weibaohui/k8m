package k8sgpt

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"k8s.io/klog/v2"
)

type K8sGPTLifecycle struct{}

func (k *K8sGPTLifecycle) Install(ctx plugins.InstallContext) error {
	klog.V(6).Infof("开始安装K8sGPT插件")
	klog.V(6).Infof("安装K8sGPT插件成功")
	return nil
}

func (k *K8sGPTLifecycle) Upgrade(ctx plugins.UpgradeContext) error {
	klog.V(6).Infof("升级K8sGPT插件：从版本 %s 到版本 %s", ctx.FromVersion(), ctx.ToVersion())
	return nil
}

func (k *K8sGPTLifecycle) Enable(ctx plugins.EnableContext) error {
	klog.V(6).Infof("启用K8sGPT插件")
	return nil
}

func (k *K8sGPTLifecycle) Disable(ctx plugins.BaseContext) error {
	klog.V(6).Infof("禁用K8sGPT插件")
	return nil
}

func (k *K8sGPTLifecycle) Uninstall(ctx plugins.UninstallContext) error {
	klog.V(6).Infof("开始卸载K8sGPT插件")
	klog.V(6).Infof("卸载K8sGPT插件成功")
	return nil
}

func (k *K8sGPTLifecycle) Start(ctx plugins.BaseContext) error {
	klog.V(6).Infof("启动K8sGPT插件成功")
	return nil
}

func (k *K8sGPTLifecycle) StartCron(ctx plugins.BaseContext, spec string) error {
	return nil
}

func (k *K8sGPTLifecycle) Stop(ctx plugins.BaseContext) error {
	klog.V(6).Infof("停止K8sGPT插件后台任务")
	return nil
}
