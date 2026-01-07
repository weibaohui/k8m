package k8m_mcp_server

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"k8s.io/klog/v2"
)

type K8mMcpServerLifecycle struct{}

func (k *K8mMcpServerLifecycle) Install(ctx plugins.InstallContext) error {
	klog.V(6).Infof("开始安装K8M MCP Server插件")
	klog.V(6).Infof("安装K8M MCP Server插件成功")
	return nil
}

func (k *K8mMcpServerLifecycle) Upgrade(ctx plugins.UpgradeContext) error {
	klog.V(6).Infof("升级K8M MCP Server插件：从版本 %s 到版本 %s", ctx.FromVersion(), ctx.ToVersion())
	return nil
}

func (k *K8mMcpServerLifecycle) Enable(ctx plugins.EnableContext) error {
	klog.V(6).Infof("启用K8M MCP Server插件")
	return nil
}

func (k *K8mMcpServerLifecycle) Disable(ctx plugins.BaseContext) error {
	klog.V(6).Infof("禁用K8M MCP Server插件")
	return nil
}

func (k *K8mMcpServerLifecycle) Uninstall(ctx plugins.UninstallContext) error {
	klog.V(6).Infof("开始卸载K8M MCP Server插件")
	klog.V(6).Infof("卸载K8M MCP Server插件成功")
	return nil
}

func (k *K8mMcpServerLifecycle) Start(ctx plugins.BaseContext) error {
	klog.V(6).Infof("启动K8M MCP Server插件成功")
	return nil
}

func (k *K8mMcpServerLifecycle) StartCron(ctx plugins.BaseContext, spec string) error {
	return nil
}

func (k *K8mMcpServerLifecycle) Stop(ctx plugins.BaseContext) error {
	klog.V(6).Infof("停止K8M MCP Server插件后台任务")
	return nil
}
