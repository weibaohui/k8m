package models

import (
	"github.com/weibaohui/k8m/internal/dao"
	"k8s.io/klog/v2"
)

func InitDB() error {
	err := dao.DB().AutoMigrate(
		&MCPServerConfig{},
		&MCPTool{},
		&MCPToolLog{},
		&McpKey{},
	)

	return err
}

func UpgradeDB(fromVersion string, toVersion string) error {
	klog.V(6).Infof("开始升级 MCP 插件数据库：从版本 %s 到版本 %s", fromVersion, toVersion)
	if err := dao.DB().AutoMigrate(
		&MCPServerConfig{},
		&MCPTool{},
		&MCPToolLog{},
		&McpKey{},
	); err != nil {
		klog.V(6).Infof("自动迁移 MCP 插件数据库失败: %v", err)
		return err
	}
	klog.V(6).Infof("升级 MCP 插件数据库完成")
	return nil
}

func DropDB() error {
	db := dao.DB()
	tables := []interface{}{
		&MCPServerConfig{},
		&MCPTool{},
		&MCPToolLog{},
		&McpKey{},
	}
	for _, t := range tables {
		if db.Migrator().HasTable(t) {
			if err := db.Migrator().DropTable(t); err != nil {
				klog.V(6).Infof("删除表失败: %v", err)
				return err
			}
		}
	}
	klog.V(6).Infof("已删除 MCP 插件表及数据")
	return nil
}
