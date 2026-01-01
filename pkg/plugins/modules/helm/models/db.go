package models

import (
	"github.com/weibaohui/k8m/internal/dao"
	"k8s.io/klog/v2"
)

// InitDB 初始化数据库表（GORM自动迁移）
func InitDB() error {
	return dao.DB().AutoMigrate(&HelmRepository{}, &HelmChart{}, &HelmRelease{})
}

// UpgradeDB 升级 Helm 插件数据库结构与数据
func UpgradeDB(fromVersion string, toVersion string) error {
	klog.V(6).Infof("开始升级 Helm 插件数据库：从版本 %s 到版本 %s", fromVersion, toVersion)
	if err := dao.DB().AutoMigrate(&HelmRepository{}, &HelmChart{}, &HelmRelease{}); err != nil {
		klog.V(6).Infof("自动迁移 Helm 插件数据库失败: %v", err)
		return err
	}
	klog.V(6).Infof("升级 Helm 插件数据库完成")
	return nil
}

// DropDB 删除 Helm 插件相关的表及数据
func DropDB() error {
	db := dao.DB()
	if db.Migrator().HasTable(&HelmRepository{}) {
		if err := db.Migrator().DropTable(&HelmRepository{}); err != nil {
			klog.V(6).Infof("删除 Helm Repository 表失败: %v", err)
			return err
		}
	}
	if db.Migrator().HasTable(&HelmChart{}) {
		if err := db.Migrator().DropTable(&HelmChart{}); err != nil {
			klog.V(6).Infof("删除 Helm Chart 表失败: %v", err)
			return err
		}
	}
	if db.Migrator().HasTable(&HelmRelease{}) {
		if err := db.Migrator().DropTable(&HelmRelease{}); err != nil {
			klog.V(6).Infof("删除 Helm Release 表失败: %v", err)
			return err
		}
	}
	klog.V(6).Infof("已删除 Helm 插件表及数据")
	return nil
}
