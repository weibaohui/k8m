package swagger

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/swagger/route"
)

var Metadata = plugins.Module{
	Meta: plugins.Meta{
		Name:        modules.PluginNameSwagger,
		Title:       "Swagger文档",
		Version:     "1.0.0",
		Description: "Swagger API文档查看",
	},
	Tables: []string{},
	Crons:  []string{},
	Menus: []plugins.Menu{
		{
			Key:   "plugin_swagger",
			Title: "Swagger文档",
			Icon:  "fa-solid fa-code-branch",
			Order: 190,
			Children: []plugins.Menu{
				{
					Key:         "plugin_swagger_open",
					Title:       "API文档",
					Icon:        "fa-solid fa-book",
					EventType:   "custom",
					CustomEvent: `() => open("/swagger/index.html")`,
					Order:       100,
				},
			},
		},
	},
	Dependencies: []string{},
	RunAfter:     []string{},

	Lifecycle:         &SwaggerLifecycle{},
	PluginAdminRouter: route.RegisterAdminRoutes,
}
