package models

import (
	"fmt"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/flag"
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	rm "github.com/weibaohui/k8m/pkg/plugins/modules/mcp_runtime/models"
	"k8s.io/klog/v2"
)

func InitDB() error {

	if plugins.ManagerInstance().IsRunning(modules.PluginNameMCPRuntime) {
		addInnerMCPServer()
	}
	return nil
}

// AddInnerMCPServer 检查并初始化名为 "k8m" 的内部 MCP 服务器配置，不存在则创建，已存在则更新其 URL。
func addInnerMCPServer() error {
	// 检查是否存在名为k8m的记录
	var count int64
	if err := dao.DB().Model(&rm.MCPServerConfig{}).Where("name = ?", "k8m").Count(&count).Error; err != nil {
		klog.Errorf("查询MCP服务器配置失败: %v", err)
		return err
	}
	cfg := flag.Init()
	// 如果不存在，添加默认的内部MCP服务器配置
	if count == 0 {
		config := &rm.MCPServerConfig{
			Name:    "k8m",
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
			Where("name =?", "k8m").
			Update("url", fmt.Sprintf("http://localhost:%d/mcp/k8m/sse", cfg.Port))
	}

	return nil
}
