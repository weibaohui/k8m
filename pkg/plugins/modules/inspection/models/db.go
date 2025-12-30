package models

import (
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/constants"
	"k8s.io/klog/v2"
)

// InitDB 初始化集群巡检相关数据库表（GORM 自动迁移），并写入内置脚本。
func InitDB() error {
	if err := dao.DB().AutoMigrate(
		&InspectionCheckEvent{},
		&InspectionRecord{},
		&InspectionSchedule{},
		&InspectionScriptResult{},
		&InspectionLuaScript{},
		&InspectionLuaScriptBuiltinVersion{},
	); err != nil {
		return err
	}
	// 初始化或升级内置脚本版本及内容
	if err := AddBuiltinLuaScripts(); err != nil {
		klog.V(6).Infof("初始化内置巡检脚本失败: %v", err)
		return err
	}
	return nil
}

// UpgradeDB 升级集群巡检插件数据库结构与数据。
// 目前仅执行 AutoMigrate，后续如有字段变更可在此处补充迁移逻辑。
func UpgradeDB(fromVersion string, toVersion string) error {
	klog.V(6).Infof("开始升级集群巡检插件数据库：从版本 %s 到版本 %s", fromVersion, toVersion)
	if err := InitDB(); err != nil {
		klog.V(6).Infof("自动迁移集群巡检插件数据库失败: %v", err)
		return err
	}
	klog.V(6).Infof("升级集群巡检插件数据库完成")
	return nil
}

// DropDB 删除集群巡检插件相关的表及数据。
func DropDB() error {
	db := dao.DB()
	// 注意：删除顺序尽量与外键依赖相反，避免约束冲突
	if db.Migrator().HasTable(&InspectionCheckEvent{}) {
		if err := db.Migrator().DropTable(&InspectionCheckEvent{}); err != nil {
			klog.V(6).Infof("删除 InspectionCheckEvent 表失败: %v", err)
			return err
		}
	}
	if db.Migrator().HasTable(&InspectionScriptResult{}) {
		if err := db.Migrator().DropTable(&InspectionScriptResult{}); err != nil {
			klog.V(6).Infof("删除 InspectionScriptResult 表失败: %v", err)
			return err
		}
	}
	if db.Migrator().HasTable(&InspectionRecord{}) {
		if err := db.Migrator().DropTable(&InspectionRecord{}); err != nil {
			klog.V(6).Infof("删除 InspectionRecord 表失败: %v", err)
			return err
		}
	}
	if db.Migrator().HasTable(&InspectionSchedule{}) {
		if err := db.Migrator().DropTable(&InspectionSchedule{}); err != nil {
			klog.V(6).Infof("删除 InspectionSchedule 表失败: %v", err)
			return err
		}
	}
	if db.Migrator().HasTable(&InspectionLuaScript{}) {
		if err := db.Migrator().DropTable(&InspectionLuaScript{}); err != nil {
			klog.V(6).Infof("删除 InspectionLuaScript 表失败: %v", err)
			return err
		}
	}
	if db.Migrator().HasTable(&InspectionLuaScriptBuiltinVersion{}) {
		if err := db.Migrator().DropTable(&InspectionLuaScriptBuiltinVersion{}); err != nil {
			klog.V(6).Infof("删除 InspectionLuaScriptBuiltinVersion 表失败: %v", err)
			return err
		}
	}
	klog.V(6).Infof("已删除集群巡检插件表及数据")
	return nil
}

// AddBuiltinLuaScripts 初始化或更新内置巡检脚本
func AddBuiltinLuaScripts() error {
	// 检查数据库中记录的内置脚本版本
	db := dao.DB()
	version, err := GetBuiltinLuaScriptsVersion(db)
	if err == nil {
		// 有记录，判断是否需要更新
		if version == BuiltinLuaScriptsVersion {
			// 版本一致，无需更新
			return nil
		}
	}
	// 版本不一致或无记录，先删除所有内置脚本
	if err := db.Where("script_type = ?", constants.LuaScriptTypeBuiltin).Delete(&InspectionLuaScript{}).Error; err != nil {
		klog.Errorf("删除旧内置巡检脚本失败: %v", err)
		return err
	}
	// 插入最新内置脚本
	if err := db.CreateInBatches(BuiltinLuaScripts, 100).Error; err != nil {
		klog.Errorf("插入内置巡检脚本失败: %v", err)
		return err
	}
	// 更新版本号
	if err := SetBuiltinLuaScriptsVersion(db, BuiltinLuaScriptsVersion); err != nil {
		klog.Errorf("更新内置脚本版本号失败: %v", err)
		return err
	}
	return nil
}
