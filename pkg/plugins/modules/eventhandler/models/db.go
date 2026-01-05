package models

import (
	"github.com/weibaohui/k8m/internal/dao"
	"k8s.io/klog/v2"
)

// InitDB 中文函数注释：初始化数据库表（GORM自动迁移）。
func InitDB() error {
	return dao.DB().AutoMigrate(&K8sEventConfig{}, &K8sEvent{}, &EventForwardSetting{})
}

// UpgradeDB 中文函数注释：升级事件转发插件数据库结构与数据。
func UpgradeDB(fromVersion string, toVersion string) error {
	klog.V(6).Infof("开始升级事件转发插件数据库：从版本 %s 到版本 %s", fromVersion, toVersion)
	if dao.DB().Migrator().HasColumn("eventhandler_event_forward_settings", "event_forward_enabled") {
		_ = dao.DB().Migrator().DropColumn("eventhandler_event_forward_settings", "event_forward_enabled")
	}
	if err := dao.DB().AutoMigrate(&K8sEventConfig{}, &K8sEvent{}, &EventForwardSetting{}); err != nil {
		klog.V(6).Infof("自动迁移事件转发插件数据库失败: %v", err)
		return err
	}
	klog.V(6).Infof("升级事件转发插件数据库完成")
	return nil
}

// DropDB 中文函数注释：删除事件转发插件相关的表及数据。
func DropDB() error {
	db := dao.DB()
	if db.Migrator().HasTable(&K8sEventConfig{}) {
		if err := db.Migrator().DropTable(&K8sEventConfig{}); err != nil {
			klog.V(6).Infof("删除事件转发插件表失败: %v", err)
			return err
		}
	}
	if db.Migrator().HasTable(&K8sEvent{}) {
		if err := db.Migrator().DropTable(&K8sEvent{}); err != nil {
			klog.V(6).Infof("删除事件转发插件表失败: %v", err)
			return err
		}
	}
	if db.Migrator().HasTable(&EventForwardSetting{}) {
		if err := db.Migrator().DropTable(&EventForwardSetting{}); err != nil {
			klog.V(6).Infof("删除事件转发插件表失败: %v", err)
			return err
		}
	}
	klog.V(6).Infof("已删除事件转发插件表及数据")
	return nil
}
