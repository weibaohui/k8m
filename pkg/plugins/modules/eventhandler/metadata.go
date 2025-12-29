package eventhandler

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/eventhandler/route"
)

var Metadata = plugins.Module{
	Meta: plugins.Meta{
		Name:        modules.PluginNameEventHandler,
		Title:       "事件转发插件",
		Version:     "1.0.0",
		Description: "K8s 事件采集、规则过滤与Webhook转发",
	},
	Tables: []string{
		"k8s_event_configs",
		"k8s_events",
		"eventhandler_event_forward_settings",
	},
	Menus: []plugins.Menu{
		{
			Key:   "plugin_eventhandler_index",
			Title: "事件转发插件",
			Icon:  "fa-solid fa-bell",
			Order: 60,
			Children: []plugins.Menu{
				{
					Key:         "plugin_eventhandler_setting",
					Title:       "事件转发参数",
					Icon:        "fa-solid fa-sliders",
					Show:        "isPlatformAdmin()==true",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/eventhandler/setting")`,
					Order:       90,
				},
				{
					Key:         "plugin_eventhandler_admin",
					Title:       "事件转发规则",
					Icon:        "fa-solid fa-plug-circle-bolt",
					Show:        "isPlatformAdmin()==true",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/eventhandler/admin")`,
					Order:       100,
				},
			},
		},
	},
	Dependencies: []string{
		modules.PluginNameLeader,
		modules.PluginNameWebhook,
	},
	Lifecycle:         &EventHandlerLifecycle{},
	PluginAdminRouter: route.RegisterPluginAdminRoutes,
}
