package gatewayapi

import (
	"context"
	"time"

	"github.com/weibaohui/k8m/pkg/plugins"
	"k8s.io/klog/v2"
)

type GatewayAPILifecycle struct {
	cancelStart context.CancelFunc
}

func (g *GatewayAPILifecycle) Install(ctx plugins.InstallContext) error {
	klog.V(6).Infof("安装GatewayAPI插件成功")
	return nil
}

func (g *GatewayAPILifecycle) Upgrade(ctx plugins.UpgradeContext) error {
	klog.V(6).Infof("升级GatewayAPI插件：从版本 %s 到版本 %s", ctx.FromVersion(), ctx.ToVersion())
	return nil
}

func (g *GatewayAPILifecycle) Enable(ctx plugins.EnableContext) error {
	klog.V(6).Infof("启用GatewayAPI插件")
	return nil
}

func (g *GatewayAPILifecycle) Disable(ctx plugins.BaseContext) error {
	klog.V(6).Infof("禁用GatewayAPI插件")
	return nil
}

func (g *GatewayAPILifecycle) Uninstall(ctx plugins.UninstallContext) error {
	klog.V(6).Infof("卸载GatewayAPI插件")
	if !ctx.KeepData() {
		klog.V(6).Infof("卸载GatewayAPI插件完成，已删除相关表及数据")
	} else {
		klog.V(6).Infof("卸载GatewayAPI插件完成，保留相关表及数据")
	}
	return nil
}

func (g *GatewayAPILifecycle) Start(ctx plugins.BaseContext) error {
	klog.V(6).Infof("启动GatewayAPI插件后台任务")

	startCtx, cancel := context.WithCancel(context.Background())
	g.cancelStart = cancel

	go func(meta plugins.Meta) {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				klog.V(6).Infof("GatewayAPI插件后台任务运行中，插件: %s，版本: %s", meta.Name, meta.Version)
			case <-startCtx.Done():
				klog.V(6).Infof("GatewayAPI 插件启动 goroutine 退出")
				return
			}
		}
	}(ctx.Meta())

	return nil
}

func (g *GatewayAPILifecycle) StartCron(ctx plugins.BaseContext, spec string) error {
	klog.V(6).Infof("执行GatewayAPI插件定时任务，表达式: %s", spec)
	return nil
}

func (g *GatewayAPILifecycle) Stop(ctx plugins.BaseContext) error {
	klog.V(6).Infof("停止GatewayAPI插件后台任务")

	if g.cancelStart != nil {
		g.cancelStart()
		g.cancelStart = nil
	}

	return nil
}
