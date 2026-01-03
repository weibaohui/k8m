package mcp

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/mcp_runtime/route"
)

var Metadata = plugins.Module{
	Meta: plugins.Meta{
		Name:        modules.PluginNameMCPRuntime,
		Title:       "MCP运行时管理插件",
		Version:     "1.0.0",
		Description: "管理大模型对话使用的MCP服务器。包括MCP服务器配置、工具管理、执行日志查看、开放MCP服务等功能。对话调用MCP时会自动添加Authorization头部，值为JWT token。",
	},
	Tables: []string{
		"mcp_server_configs",
		"mcp_tools",
		"mcp_tool_logs",
		"mcp_keys",
	},
	Menus: []plugins.Menu{
		{
			Key:   "plugin_mcp_index",
			Title: "MCP运行管理",
			Icon:  "fa-solid fa-network-wired",
			Order: 45,
			Children: []plugins.Menu{
				{
					Key:         "plugin_mcp_server",
					Title:       "MCP服务管理",
					Icon:        "fa-solid fa-server",
					Show:        "isPlatformAdmin()==true",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/mcp_runtime/server")`,
					Order:       10,
				},
				{
					Key:         "plugin_mcp_log",
					Title:       "MCP执行日志",
					Icon:        "fa-solid fa-list-alt",
					Show:        "isPlatformAdmin()==true",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/mcp_runtime/log")`,
					Order:       20,
				},
				{
					Key:         "plugin_mcp_keys",
					Title:       "开放MCP服务",
					Icon:        "fa-solid fa-key",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/mcp_runtime/keys")`,
					Order:       30,
				},
			},
		},
	},
	Dependencies:      []string{},
	RunAfter:          []string{modules.PluginNameK8mMcpServer},
	Lifecycle:         &McpLifecycle{},
	PluginAdminRouter: route.RegisterPluginAdminRoutes,
	ClusterRouter:     nil,
	ManagementRouter:  route.RegisterPluginMgmRoutes,
}
