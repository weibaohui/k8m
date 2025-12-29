package models

import (
	"github.com/weibaohui/k8m/internal/dao"
	"k8s.io/klog/v2"
)

// InitDB 中文函数注释：初始化数据库表（GORM自动迁移）。
func InitDB() error {
	return dao.DB().AutoMigrate(&WebhookReceiver{}, &WebhookLogRecord{})
}

// UpgradeDB 中文函数注释：升级webhook插件数据库结构与数据。
func UpgradeDB(fromVersion string, toVersion string) error {
	klog.V(6).Infof("开始升级webhook插件数据库：从版本 %s 到版本 %s", fromVersion, toVersion)
	if err := dao.DB().AutoMigrate(&WebhookReceiver{}, &WebhookLogRecord{}); err != nil {
		klog.V(6).Infof("自动迁移webhook插件数据库失败: %v", err)
		return err
	}
	klog.V(6).Infof("升级webhook插件数据库完成")
	return nil
}

// DropDB 中文函数注释：删除webhook插件相关的表及数据。
func DropDB() error {
	db := dao.DB()
	if db.Migrator().HasTable(&WebhookReceiver{}) {
		if err := db.Migrator().DropTable(&WebhookReceiver{}); err != nil {
			klog.V(6).Infof("删除webhook插件表失败: %v", err)
			return err
		}
	}
	if db.Migrator().HasTable(&WebhookLogRecord{}) {
		if err := db.Migrator().DropTable(&WebhookLogRecord{}); err != nil {
			klog.V(6).Infof("删除webhook插件表失败: %v", err)
			return err
		}
	}
	klog.V(6).Infof("已删除webhook插件表及数据")
	return nil
}
