package heartbeat

import (
	"context"

	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/eventbus"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	heartbeatinterface "github.com/weibaohui/k8m/pkg/plugins/modules/heartbeat/interface"
	"github.com/weibaohui/k8m/pkg/plugins/modules/heartbeat/models"
	"github.com/weibaohui/k8m/pkg/plugins/modules/heartbeat/service"
	gservice "github.com/weibaohui/k8m/pkg/service"
	"k8s.io/klog/v2"
)

// HeartbeatLifecycle 心跳插件生命周期实现
type HeartbeatLifecycle struct {
	manager           *service.HeartbeatManager
	leaderWatchCancel context.CancelFunc
}

// Install 安装心跳插件
func (h *HeartbeatLifecycle) Install(ctx plugins.InstallContext) error {
	if err := models.InitDB(); err != nil {
		klog.V(6).Infof("安装心跳插件失败: %v", err)
		return err
	}
	klog.V(6).Infof("安装心跳插件成功")
	return nil
}

// Upgrade 升级心跳插件
func (h *HeartbeatLifecycle) Upgrade(ctx plugins.UpgradeContext) error {
	klog.V(6).Infof("升级心跳插件：从版本 %s 到版本 %s", ctx.FromVersion(), ctx.ToVersion())
	if err := models.UpgradeDB(ctx.FromVersion(), ctx.ToVersion()); err != nil {
		klog.V(6).Infof("升级心跳插件失败: %v", err)
		return err
	}
	return nil
}

// Enable 启用心跳插件
func (h *HeartbeatLifecycle) Enable(ctx plugins.EnableContext) error {
	klog.V(6).Infof("启用心跳插件")
	return nil
}

// Disable 禁用心跳插件
func (h *HeartbeatLifecycle) Disable(ctx plugins.BaseContext) error {
	klog.V(6).Infof("禁用心跳插件")

	if h.leaderWatchCancel != nil {
		h.leaderWatchCancel()
		h.leaderWatchCancel = nil
	}

	h.StopHeartbeat()
	return nil
}

// Uninstall 卸载心跳插件
func (h *HeartbeatLifecycle) Uninstall(ctx plugins.UninstallContext) error {
	klog.V(6).Infof("卸载心跳插件")

	if h.leaderWatchCancel != nil {
		h.leaderWatchCancel()
		h.leaderWatchCancel = nil
	}

	h.StopHeartbeat()

	// 根据keepData参数决定是否删除数据库
	if !ctx.KeepData() {
		if err := models.DropDB(); err != nil {
			klog.V(6).Infof("卸载心跳插件失败: %v", err)
			return err
		}
	}
	klog.V(6).Infof("卸载心跳插件成功")
	return nil
}

// Start 启动心跳插件的后台任务
func (h *HeartbeatLifecycle) Start(ctx plugins.BaseContext) error {
	klog.V(6).Infof("启动心跳插件后台任务")

	if plugins.ManagerInstance().IsEnabled(modules.PluginNameLeader) {
		elect := ctx.Bus().Subscribe(eventbus.EventLeaderElected)
		lost := ctx.Bus().Subscribe(eventbus.EventLeaderLost)

		leaderWatchCtx, cancel := context.WithCancel(context.Background())
		h.leaderWatchCancel = cancel

		go func() {
			for {
				select {
				case <-elect:
					h.StartHeartbeat()
					klog.V(6).Infof("成为Leader，启动心跳检测")
				case <-lost:
					h.StopHeartbeat()
					klog.V(6).Infof("不再是Leader，停止心跳检测")
				case <-leaderWatchCtx.Done():
					klog.V(6).Infof("心跳插件 Leader 监听 goroutine 退出")
					return
				}
			}
		}()
		klog.V(6).Infof("根据实例Leader获取状态启动事件转发插件后台任务")
	} else {
		h.StartHeartbeat()
		klog.V(6).Infof("启动心跳插件后台任务")
	}

	return nil
}

// StartCron 启动心跳插件的定时任务
func (h *HeartbeatLifecycle) StartCron(ctx plugins.BaseContext, spec string) error {
	klog.V(6).Infof("执行心跳插件定时任务，表达式: %s", spec)
	return nil
}

func (h *HeartbeatLifecycle) StartHeartbeat() {
	// 初始化心跳管理器
	h.manager = service.NewHeartbeatManager()
	// 设置全局实例，以便主服务可以访问
	heartbeatinterface.GlobalHeartbeatManager = h.manager

	// 为所有集群启动心跳检测
	clusters := gservice.ClusterService().AllClusters()
	for _, cluster := range clusters {
		if cluster.ClusterConnectStatus == "connected" {
			h.manager.StartHeartbeat(cluster.ClusterID)
		}
	}

	klog.V(6).Infof("心跳插件已启用，为 %d 个已连接集群启动了心跳检测", len(clusters))
}

func (h *HeartbeatLifecycle) StopHeartbeat() {
	if h.manager != nil {
		// 为所有集群停止心跳检测和自动重连
		clusters := gservice.ClusterService().AllClusters()
		for _, cluster := range clusters {
			h.manager.StopHeartbeat(cluster.ClusterID)
			h.manager.StopReconnect(cluster.ClusterID)
		}
		klog.V(6).Infof("已为 %d 个集群停止心跳检测和自动重连", len(clusters))
	}
}
