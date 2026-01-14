package models

import (
	"errors"
	"fmt"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/flag"
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	rm "github.com/weibaohui/k8m/pkg/plugins/modules/mcp_runtime/models"
	"github.com/weibaohui/k8m/pkg/plugins/modules/mcp_runtime/service"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

const innerMCPServerName = "k8m"

// InitDB 初始化本插件安装时需要写入的数据库数据。
func InitDB() error {
	if plugins.ManagerInstance().IsRunning(modules.PluginNameMCPRuntime) {
		if err := addInnerMCPServer(); err != nil {
			klog.V(6).Infof("初始化内置 MCP Server 配置失败: %v", err)
			return err
		}
	}
	return nil
}

// addInnerMCPServer 检查并初始化名为 "k8m" 的内部 MCP 服务器配置，不存在则创建，已存在则更新其 URL。
func addInnerMCPServer() error {
	// 检查是否存在名为k8m的记录
	var count int64
	if err := dao.DB().Model(&rm.MCPServerConfig{}).Where("name = ?", innerMCPServerName).Count(&count).Error; err != nil {
		klog.Errorf("查询MCP服务器配置失败: %v", err)
		return err
	}
	cfg := flag.Init()
	// 如果不存在，添加默认的内部MCP服务器配置
	if count == 0 {
		config := &rm.MCPServerConfig{
			Name:    innerMCPServerName,
			URL:     fmt.Sprintf("http://localhost:%d/mcp/k8m/sse", cfg.Port),
			Enabled: false,
		}
		if err := dao.DB().Create(config).Error; err != nil {
			klog.Errorf("添加内部MCP服务器配置失败: %v", err)
			return err
		}
		klog.V(4).Info("成功添加内部MCP服务器配置")
	} else {
		klog.V(4).Info("内部MCP服务器配置已存在")
		dao.DB().Model(&rm.MCPServerConfig{}).Select("url").
			Where("name =?", innerMCPServerName).
			Update("url", fmt.Sprintf("http://localhost:%d/mcp/k8m/sse", cfg.Port))
	}

	return nil
}

// DropDB 在卸载本插件且不保留数据时，删除内置 MCP Server 配置及相关数据。
func DropDB() error {
	db := dao.DB()

	if !db.Migrator().HasTable(&rm.MCPServerConfig{}) {
		klog.V(6).Infof("未发现 MCP Server 配置表，跳过删除内置服务器配置[%s]", innerMCPServerName)
		return nil
	}

	var server rm.MCPServerConfig
	if err := db.Where("name = ?", innerMCPServerName).First(&server).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			klog.V(6).Infof("未找到内置 MCP Server 配置[%s]，无需删除", innerMCPServerName)
			return nil
		}
		klog.V(6).Infof("查询内置 MCP Server 配置失败[%s]: %v", innerMCPServerName, err)
		return err
	}

	if db.Migrator().HasTable(&rm.MCPTool{}) {
		if err := db.Where("server_name = ?", innerMCPServerName).Delete(&rm.MCPTool{}).Error; err != nil {
			klog.V(6).Infof("删除内置 MCP Server 工具记录失败[%s]: %v", innerMCPServerName, err)
			return err
		}
	}
	if db.Migrator().HasTable(&rm.MCPToolLog{}) {
		if err := db.Where("server_name = ?", innerMCPServerName).Delete(&rm.MCPToolLog{}).Error; err != nil {
			klog.V(6).Infof("删除内置 MCP Server 工具日志失败[%s]: %v", innerMCPServerName, err)
			return err
		}
	}

	if err := db.Where("name = ?", innerMCPServerName).Delete(&rm.MCPServerConfig{}).Error; err != nil {
		klog.V(6).Infof("删除内置 MCP Server 配置失败[%s]: %v", innerMCPServerName, err)
		return err
	}

	if plugins.ManagerInstance().IsRunning(modules.PluginNameMCPRuntime) {
		service.McpService().RemoveServer(server)
	}

	klog.V(6).Infof("已删除内置 MCP Server 配置及相关数据[%s]", innerMCPServerName)
	return nil
}
