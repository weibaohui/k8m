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
	RouteRules: []plugins.RouteRule{
		{Method: "GET", Path: "/items", Kind: plugins.AccessAnyUser},
		{Method: "POST", Path: "/items", Kind: plugins.AccessPlatformAdmin},
		{Method: "POST", Path: "/items/:id", Kind: plugins.AccessPlatformAdmin},
		{Method: "POST", Path: "/remove/items/:id", Kind: plugins.AccessPlatformAdmin},
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
	Permissions: []plugins.Permission{
		{Name: "demo.view", Title: "查看演示列表"},
		{Name: "demo.edit", Title: "编辑演示列表"},
	},

	Lifecycle: &DemoLifecycle{},
	Router:    backend.RegisterRoutes,
}
