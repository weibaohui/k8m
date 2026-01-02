package mcp

import (
	"context"
	"sync"
	"time"

	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules/mcp/models"
	"github.com/weibaohui/k8m/pkg/plugins/modules/mcp/service"
	"k8s.io/klog/v2"
)

type McpLifecycle struct {
	mu            sync.Mutex
	leaderWatchMu sync.Mutex
	stopChan      chan struct{}
	cancelLeader  context.CancelFunc
}

func (l *McpLifecycle) Install(ctx plugins.InstallContext) error {
	if err := models.InitDB(); err != nil {
		klog.V(6).Infof("安装 MCP 插件失败: %v", err)
		return err
	}
	klog.V(6).Infof("安装 MCP 插件成功")
	return nil
}

func (l *McpLifecycle) Upgrade(ctx plugins.UpgradeContext) error {
	klog.V(6).Infof("升级 MCP 插件：从版本 %s 到版本 %s", ctx.FromVersion(), ctx.ToVersion())
	if err := models.UpgradeDB(ctx.FromVersion(), ctx.ToVersion()); err != nil {
		klog.V(6).Infof("升级 MCP 插件失败: %v", err)
		return err
	}
	return nil
}

func (l *McpLifecycle) Enable(ctx plugins.EnableContext) error {
	klog.V(6).Infof("启用 MCP 插件")
	return nil
}

func (l *McpLifecycle) Disable(ctx plugins.BaseContext) error {
	klog.V(6).Infof("禁用 MCP 插件")
	return nil
}

func (l *McpLifecycle) Uninstall(ctx plugins.UninstallContext) error {
	if !ctx.KeepData() {
		if err := models.DropDB(); err != nil {
			klog.V(6).Infof("卸载 MCP 插件失败: %v", err)
			return err
		}
	}
	klog.V(6).Infof("卸载 MCP 插件成功")
	return nil
}

func (l *McpLifecycle) Start(ctx plugins.BaseContext) error {
	service.McpService().Init()

	//todo MCPService 启动完毕后，再start，使用sleep 不好，最好改成事件性质的
	time.Sleep(120 * time.Second)
	service.McpService().Start()

	klog.V(6).Infof("MCP 插件已启动")
	return nil
}

func (l *McpLifecycle) StartCron(ctx plugins.BaseContext, spec string) error {
	return nil
}
