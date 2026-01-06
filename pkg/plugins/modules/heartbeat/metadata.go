package heartbeat

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/heartbeat/route"
)

var Metadata = plugins.Module{
	Meta: plugins.Meta{
		Name:        modules.PluginNameHeartbeat,
		Title:       "集群心跳检测",
		Version:     "1.0.0",
		Description: "管理集群心跳检测和自动重连功能",
	},
	Menus: []plugins.Menu{
		{
			Key:   "plugin_heartbeat_index",
			Title: "心跳检测",
			Icon:  "fa-solid fa-heartbeat",
			Order: 15,
			Children: []plugins.Menu{
				{
					Key:         "plugin_heartbeat_config",
					Title:       "心跳配置",
					Icon:        "fa-solid fa-gear",
					Show:        "isPlatformAdmin()==true",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/heartbeat/config")`,
					Order:       100,
				},
			},
		},
	},
	Tables: []string{
		"heartbeat_settings",
	},
	Dependencies: []string{},
	RunAfter: []string{
		modules.PluginNameLeader,
	},

	Lifecycle: &HeartbeatLifecycle{},
	// 插件管理员类操作API，要求是平台管理员，用于心跳配置
	PluginAdminRouter: route.RegisterPluginAdminRoutes,
}
