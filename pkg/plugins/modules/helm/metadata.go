package helm

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/helm/route"
)

var Metadata = plugins.Module{
	Meta: plugins.Meta{
		Name:        modules.PluginNameHelm,
		Title:       "Helm 管理插件",
		Version:     "1.0.0",
		Description: "Helm 仓库、Chart、Release 管理。包括仓库添加、Chart浏览、Release安装升级卸载等功能。定时更新仓库索引。",
	},
	Tables: []string{
		"helm_repositories",
		"helm_charts",
		"helm_releases",
	},
	Menus: []plugins.Menu{
		{
			Key:   "plugin_helm_index",
			Title: "Helm 管理",
			Icon:  "fa-solid fa-ship",
			Order: 50,
			Children: []plugins.Menu{
				{
					Key:         "plugin_helm_repo",
					Title:       "Helm 仓库管理",
					Icon:        "fa-solid fa-warehouse",
					Show:        "isPlatformAdmin()==true",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/helm/repo")`,
					Order:       10,
				},
				{
					Key:         "plugin_helm_run_params",
					Title:       "Helm 运行参数管理",
					Icon:        "fa-solid fa-gear",
					Show:        "isPlatformAdmin()==true",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/helm/setting")`,
					Order:       10,
				},
				{
					Key:         "plugin_helm_chart",
					Title:       "Chart 浏览",
					Icon:        "fa-solid fa-cube",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/helm/chart")`,
					Order:       20,
				},
				{
					Key:         "plugin_helm_release",
					Title:       "Release 管理",
					Icon:        "fa-solid fa-rocket",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/helm/release")`,
					Order:       30,
				},
			},
		},
	},
	Dependencies:      []string{},
	RunAfter:          []string{modules.PluginNameLeader},
	Lifecycle:         &HelmLifecycle{},
	PluginAdminRouter: route.RegisterPluginAdminRoutes,
	ClusterRouter:     route.RegisterPluginAPIRoutes,
	ManagementRouter:  route.RegisterPluginMgmRoutes,
}
