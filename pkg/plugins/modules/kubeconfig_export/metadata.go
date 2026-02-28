package kubeconfig_export

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/kubeconfig_export/route"
)

var Metadata = plugins.Module{
	Meta: plugins.Meta{
		Name:        modules.PluginNameKubeconfigExport,
		Title:       "Kubeconfig 导出",
		Version:     "1.0.0",
		Description: "为集群生成 Kubeconfig 并提供导出功能",
	},
	Tables: []string{
		// 本插件不创建额外的数据库表，使用现有 kube_configs 表
	},
	Menus: []plugins.Menu{
		{
			Key:   "plugin_kubeconfig_export_index",
			Title: "Kubeconfig 导出",
			Icon:  "fa-solid fa-download",
			Order: 10,
			Children: []plugins.Menu{
				{
					Key:         "plugin_kubeconfig_export_mgm",
					Title:       "Kubeconfig 管理",
					Icon:        "fa-solid fa-file-export",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/kubeconfig_export/management")`,
					Order:       100,
				},
			},
		},
	},
	Dependencies: []string{},
	RunAfter: []string{
		modules.PluginNameLeader,
	},
	Lifecycle:     &KubeconfigExportLifecycle{},
	ClusterRouter: route.RegisterClusterRoutes,
	// 管理类操作路由
	ManagementRouter: route.RegisterManagementRoutes,
	// 插件管理员类路由（可选）
	PluginAdminRouter: route.RegisterPluginAdminRoutes,
}