package inspection

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/inspection/route"
)

// Metadata 巡检插件元信息与能力声明
var Metadata = plugins.Module{
	Meta: plugins.Meta{
		Name:        modules.PluginNameInspection,
		Title:       "集群巡检插件",
		Version:     "1.0.0",
		Description: "基于 Lua 的集群巡检计划、规则管理与结果查看",
	},
	// 相关数据表名称，仅用于展示和管理，无强约束
	Tables: []string{
		"inspection_schedules",
		"inspection_records",
		"inspection_check_events",
		"inspection_script_results",
		"inspection_lua_scripts",
		"inspection_lua_script_builtin_versions",
	},
	// 菜单声明：这里直接复用已有的 /admin/inspection 下的页面
	Menus: []plugins.Menu{
		{
			Key:   "plugin_inspection_index",
			Title: "集群巡检",
			Icon:  "fa-solid fa-stethoscope",
			Order: 40,
			Children: []plugins.Menu{
				{
					Key:         "plugin_inspection_schedule",
					Title:       "巡检计划",
					Icon:        "fa-regular fa-calendar-check",
					EventType:   "custom",
					CustomEvent: "() => loadJsonPage(\"/admin/inspection/schedule\")",
					Order:       100,
				},
				{
					Key:         "plugin_inspection_script",
					Title:       "巡检规则",
					Icon:        "fa-solid fa-code",
					EventType:   "custom",
					CustomEvent: "() => loadJsonPage(\"/admin/inspection/script\")",
					Order:       101,
				},
				{
					Key:         "plugin_inspection_lua_doc",
					Title:       "Lua 规则说明",
					Icon:        "fa-regular fa-file-lines",
					EventType:   "custom",
					CustomEvent: "() => loadJsonPage(\"/admin/inspection/lua_doc\")",
					Order:       102,
				},
			},
		},
	},
	// 依赖：需要 leader 插件提供主备能力，以及 webhook 插件用于通知
	Dependencies: []string{
		modules.PluginNameLeader,
		modules.PluginNameWebhook,
	},
	Lifecycle:         &InspectionLifecycle{},
	PluginAdminRouter: route.RegisterPluginAdminRoutes,
}
