package mcp

import (
	"context"
	"sync"
	"time"

	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules/mcp_runtime/models"
	"github.com/weibaohui/k8m/pkg/plugins/modules/mcp_runtime/service"
	"k8s.io/klog/v2"
)

type McpLifecycle struct {
	mu            sync.Mutex
	leaderWatchMu sync.Mutex
	stopChan      chan struct{}
	cancelLeader  context.CancelFunc
	cancelStart   context.CancelFunc
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

	if l.cancelStart != nil {
		l.cancelStart()
		l.cancelStart = nil
	}

	return nil
}

func (l *McpLifecycle) Uninstall(ctx plugins.UninstallContext) error {

	if l.cancelStart != nil {
		l.cancelStart()
		l.cancelStart = nil
	}

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

	startCtx, cancel := context.WithCancel(context.Background())
	l.cancelStart = cancel

	go func() {
		service.McpService().Init()
		select {
		case <-time.After(30 * time.Second):
			service.McpService().Start()
		case <-startCtx.Done():
			klog.V(6).Infof("MCP 插件启动 goroutine 退出")
			return
		}
	}()

	klog.V(6).Infof("MCP 插件已启动")
	return nil
}

func (l *McpLifecycle) StartCron(ctx plugins.BaseContext, spec string) error {
	return nil
}
