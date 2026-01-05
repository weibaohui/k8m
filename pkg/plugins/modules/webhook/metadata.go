package webhook

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/webhook/route"
)

var Metadata = plugins.Module{
	Meta: plugins.Meta{
		Name:        modules.PluginNameWebhook,
		Title:       "Webhook插件",
		Version:     "1.0.0",
		Description: "Webhook接收器管理、测试发送与发送记录查询",
	},
	Tables: []string{
		"webhook_receiver",
		"webhook_log_record",
	},
	Menus: []plugins.Menu{
		{
			Key:   "plugin_webhook_index",
			Title: "Webhook插件",
			Icon:  "fa-solid fa-link",
			Order: 50,
			Children: []plugins.Menu{
				{
					Key:         "plugin_webhook_admin",
					Title:       "Webhook管理",
					Icon:        "fa-solid fa-gear",
					Show:        "isPlatformAdmin()==true",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/webhook/admin")`,
					Order:       100,
				},
				{
					Key:         "plugin_webhook_records",
					Title:       "Webhook记录",
					Icon:        "fa-solid fa-list",
					Show:        "isPlatformAdmin()==true",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/webhook/records")`,
					Order:       101,
				},
			},
		},
	},
	Lifecycle:         &WebhookLifecycle{},
	PluginAdminRouter: route.RegisterPluginAdminRoutes,
}
