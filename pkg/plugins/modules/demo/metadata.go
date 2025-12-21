package demo

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules/demo/backend"
)

var Metadata = plugins.Module{
	Meta: plugins.Meta{
		Name:        "demo",
		Title:       "演示插件",
		Version:     "1.0.12",
		Description: "演示固定列表与CRUD功能",
	},
	Menus: []plugins.Menu{
		{
			Key:         "plugin_demo",
			Title:       "演示插件",
			Icon:        "fa-solid fa-puzzle-piece",
			EventType:   "custom",
			CustomEvent: `() => loadJsonPage("/plugins/demo/page")`,
			Order:       100,
			Permission:  "demo.view",
		},
	},

	Lifecycle: &DemoLifecycle{},
	Router:    backend.RegisterRoutes,
}
