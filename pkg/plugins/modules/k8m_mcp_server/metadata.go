package k8m_mcp_server

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/k8m_mcp_server/route"
)

var Metadata = plugins.Module{
	Meta: plugins.Meta{
		Name:        modules.PluginNameK8mMcpServer,
		Title:       "K8M MCP Server插件",
		Version:     "1.0.0",
		Description: "将K8M作为MCP Server使用。可以添加到MCP运行管理中使用。",
	},
	Tables:            []string{},
	Menus:             []plugins.Menu{},
	Dependencies:      []string{},
	RunAfter:          []string{},
	Lifecycle:         &K8mMcpServerLifecycle{},
	PluginAdminRouter: nil,
	ClusterRouter:     nil,
	RootRouter:        route.RegisterRootRoutes,
}
