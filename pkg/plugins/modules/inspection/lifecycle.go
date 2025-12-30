package inspection

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/inspection/lua"
	"github.com/weibaohui/k8m/pkg/plugins/modules/inspection/models"
	"github.com/weibaohui/k8m/pkg/service"
	"k8s.io/klog/v2"
)

// InspectionLifecycle 巡检插件生命周期实现
// 数据库迁移由插件自身负责（通过 InitDB/UpgradeDB），巡检任务调度则由 leader 插件在成为 Leader 时按插件状态调用 lua.InitClusterInspection 完成。
type InspectionLifecycle struct{}

func (l *InspectionLifecycle) Install(ctx plugins.InstallContext) error {
	klog.V(6).Infof("安装集群巡检插件")
	if err := models.InitDB(); err != nil {
		klog.V(6).Infof("安装集群巡检插件失败，初始化数据库失败: %v", err)
		return err
	}
	return nil
}

func (l *InspectionLifecycle) Upgrade(ctx plugins.UpgradeContext) error {
	klog.V(6).Infof("升级集群巡检插件：从版本 %s 到版本 %s", ctx.FromVersion(), ctx.ToVersion())
	if err := models.UpgradeDB(ctx.FromVersion(), ctx.ToVersion()); err != nil {
		klog.V(6).Infof("升级集群巡检插件失败: %v", err)
		return err
	}
	return nil
}

func (l *InspectionLifecycle) Enable(ctx plugins.EnableContext) error {
	// 启用时确保表结构存在
	if err := models.InitDB(); err != nil {
		klog.V(6).Infof("启用集群巡检插件失败: %v", err)
		return err
	}
	klog.V(6).Infof("启用集群巡检插件")
	return nil
}

func (l *InspectionLifecycle) Disable(ctx plugins.BaseContext) error {
	klog.V(6).Infof("禁用集群巡检插件")
	return nil
}

func (l *InspectionLifecycle) Uninstall(ctx plugins.UninstallContext) error {
	klog.V(6).Infof("卸载集群巡检插件")
	// 根据 keepData 参数决定是否删除表结构
	if !ctx.KeepData() {
		if err := models.DropDB(); err != nil {
			klog.V(6).Infof("卸载集群巡检插件失败: %v", err)
			return err
		}
	}
	return nil
}

// Start 集群巡检插件自身没有需要在所有实例上运行的后台任务
// 真正的定时巡检调度由 leader 插件在成为 Leader 时统一启动
func (l *InspectionLifecycle) Start(ctx plugins.BaseContext) error {
	//启用了 leader 插件，且当前实例是 Leader 时才初始化巡检任务
	if plugins.ManagerInstance().IsEnabled(modules.PluginNameLeader) {
		if service.LeaderService().IsCurrentLeader() {
			lua.InitClusterInspection()
		}
	} else {
		lua.InitClusterInspection()
	}
	return nil
}

// StartCron 当前巡检插件不使用插件级 cron 表达式
func (l *InspectionLifecycle) StartCron(ctx plugins.BaseContext, spec string) error {
	return nil
}
