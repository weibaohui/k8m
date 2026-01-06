package k8sgpt

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/k8sgpt/route"
)

var Metadata = plugins.Module{
	Meta: plugins.Meta{
		Name:        modules.PluginNameK8sGPT,
		Title:       "K8sGPT插件",
		Version:     "1.0.0",
		Description: "Kubernetes资源AI智能分析，支持Pod、Deployment、Service等多种资源类型的智能诊断。源自https://github.com/k8sgpt-ai/k8sgpt项目",
	},
	Tables: []string{},
	Menus: []plugins.Menu{
		{
			Key:   "plugin_k8sgpt_index",
			Title: "K8sGPT",
			Icon:  "fa-solid fa-brain",
			Order: 40,
			Children: []plugins.Menu{
				{
					Key:         "plugin_k8sgpt_analysis",
					Title:       "资源分析",
					Icon:        "fa-solid fa-magnifying-glass-chart",
					Show:        "true",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/k8sgpt/analysis")`,
					Order:       100,
				},
			},
		},
	},
	Dependencies:     []string{modules.PluginNameAI},
	Lifecycle:        &K8sGPTLifecycle{},
	ClusterRouter:    route.RegisterClusterRoutes,
	ManagementRouter: route.RegisterMgmRoutes,
}
