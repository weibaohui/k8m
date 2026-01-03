package openapi

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/openapi/route"
)

var Metadata = plugins.Module{
	Meta: plugins.Meta{
		Name:        modules.PluginNameOpenAPI,
		Title:       "OpenAPI插件",
		Version:     "1.0.0",
		Description: "API密钥管理，用于程序化访问平台",
	},
	Tables: []string{
		"api_keys",
	},
	Menus: []plugins.Menu{
		{
			Key:   "plugin_openapi_index",
			Title: "OpenAPI",
			Icon:  "fa-solid fa-key",
			Order: 60,
			Children: []plugins.Menu{
				{
					Key:         "plugin_openapi_keys",
					Title:       "API密钥",
					Icon:        "fa-solid fa-key",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/openapi/api_keys")`,
					Order:       100,
				},
			},
		},
	},
	Lifecycle:        &OpenAPILifecycle{},
	ManagementRouter: route.RegisterPluginMgmRoutes,
}
