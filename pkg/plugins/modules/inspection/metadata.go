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
		Description: "基于 Lua 的集群巡检计划、规则管理与结果查看。启用选举插件后，只有主实例执行，否则每个实例都执行。",
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
	// 菜单声明：使用插件专属路径
	Menus: []plugins.Menu{
		{
			Key:   "plugin_inspection_index",
			Title: "集群巡检插件",
			Icon:  "fa-solid fa-stethoscope",
			Order: 40,
			Show:  "isPlatformAdmin()==true",
			Children: []plugins.Menu{
				{
					Key:         "plugin_inspection_summary",
					Title:       "巡检汇总",
					Icon:        "fa-solid fa-clipboard-list",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/inspection/summary")`,
					Order:       1,
				},
				{
					Key:         "plugin_inspection_schedule",
					Title:       "巡检计划",
					Icon:        "fa-regular fa-calendar-check",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/inspection/schedule")`,
					Order:       100,
					Show:        "isPlatformAdmin()==true",
				},
				{
					Key:         "plugin_inspection_script",
					Title:       "巡检规则",
					Icon:        "fa-solid fa-code",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/inspection/script")`,
					Order:       101,
					Show:        "isPlatformAdmin()==true",
				},
				{
					Key:         "plugin_inspection_record",
					Title:       "巡检记录",
					Icon:        "fa-solid fa-list-check",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/inspection/record")`,
					Order:       102,
					Show:        "isPlatformAdmin()==true",
				},
				{
					Key:         "plugin_inspection_lua_doc",
					Title:       "Lua 规则说明",
					Icon:        "fa-regular fa-file-lines",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/inspection/lua_doc")`,
					Order:       103,
					Show:        "isPlatformAdmin()==true",
				},
			},
		},
	},
	// 依赖：需要 leader 插件提供主备能力，以及 webhook 插件用于通知
	Dependencies: []string{
		modules.PluginNameWebhook,
	},
	RunAfter:          []string{modules.PluginNameLeader, modules.PluginNameAI},
	Lifecycle:         &InspectionLifecycle{},
	PluginAdminRouter: route.RegisterPluginAdminRoutes,
}
