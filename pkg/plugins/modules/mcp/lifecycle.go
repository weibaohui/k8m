package mcp

import (
	"context"
	"sync"

	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/eventbus"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/mcp/models"
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
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.stopChan == nil {
		l.stopChan = make(chan struct{})
	}

	if plugins.ManagerInstance().IsEnabled(modules.PluginNameLeader) {
		elect := ctx.Bus().Subscribe(eventbus.EventLeaderElected)
		lost := ctx.Bus().Subscribe(eventbus.EventLeaderLost)

		go func() {
			for {
				select {
				case <-elect:
					klog.V(6).Infof("成为 Leader，启动 MCP 服务")
					mcpservice.StartMCPService()
				case <-lost:
					klog.V(6).Infof("不再是 Leader，停止 MCP 服务")
					mcpservice.StopMCPService()
				case <-l.stopChan:
					return
				}
			}
		}()

		klog.V(6).Infof("根据实例 Leader 状态启动 MCP 插件后台任务")
	} else {
		mcpservice.StartMCPService()
		klog.V(6).Infof("启动 MCP 插件后台任务")
	}
	return nil
}

func (l *McpLifecycle) StartCron(ctx plugins.BaseContext, spec string) error {
	return nil
}
