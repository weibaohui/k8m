package demo

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules/demo/route"
)

var Metadata = plugins.Module{
	Meta: plugins.Meta{
		Name:        "demo",
		Title:       "演示插件",
		Version:     "1.0.12",
		Description: "演示固定列表与CRUD功能",
	},
	Crons: []string{
		"* * * * *",
		"*/2 * * * *",
	},
	Menus: []plugins.Menu{
		{
			Key:   "plugin_demo_index",
			Title: "演示插件",
			Icon:  "fa-solid fa-cube",
			Order: 1,
			Children: []plugins.Menu{
				{
					Key:         "plugin_demo_cluster",
					Title:       "演示插件Cluster",
					Icon:        "fa-solid fa-puzzle-piece",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/demo/cluster")`,
					Order:       100,
				},
				{
					Key:         "plugin_demo_mgm",
					Title:       "演示插件Mgm",
					Icon:        "fa-solid fa-puzzle-piece",
					Show:        "isUserInGroup('特尔是')",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/demo/mgm")`,
					Order:       101,
				},
				{
					Key:         "plugin_demo_admin",
					Title:       "演示插件Admin",
					Icon:        "fa-solid fa-puzzle-piece",
					Show:        "isPlatformAdmin()==true",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/demo/admin")`,
					Order:       102,
				},
			},
		},
	},
	// Dependencies 插件依赖的其他插件名称列表；启用前需确保均已启用
	Dependencies: []string{
		"leader",
	},

	Lifecycle:         &DemoLifecycle{},
	ClusterRouter:     route.RegisterClusterRoutes,
	ManagementRouter:  route.RegisterManagementRoutes,
	PluginAdminRouter: route.RegisterPluginAdminRoutes,
}
