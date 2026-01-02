package gllog

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/gllog/route"
)

var Metadata = plugins.Module{
	Meta: plugins.Meta{
		Name:        modules.PluginNameGlobalLog,
		Title:       "全局日志",
		Version:     "1.0.0",
		Description: "全局日志查询，支持跨集群Pod日志查看",
	},
	Tables: []string{},
	Crons:  []string{},
	Menus: []plugins.Menu{
		{
			Key:   "plugin_gllog",
			Title: "全局日志",
			Icon:  "fa-solid fa-file-lines",
			Order: 150,
			Children: []plugins.Menu{
				{
					Key:         "plugin_gllog_query",
					Title:       "日志查询",
					Icon:        "fa-solid fa-search",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/gllog/log")`,
					Order:       100,
				},
			},
		},
	},
	Dependencies: []string{},
	RunAfter: []string{
		modules.PluginNameLeader,
	},

	Lifecycle: &GlobalLogLifecycle{},
	//管理类操作API，要求是登录用户，可用于各类操作
	ManagementRouter: route.RegisterManagementRoutes,
}
