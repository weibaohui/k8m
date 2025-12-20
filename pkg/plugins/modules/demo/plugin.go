package demo

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules/demo/backend"
)

var ModuleDef = plugins.Module{
	Meta: plugins.Meta{
		Name:        "demo",
		Title:       "演示插件",
		Version:     "1.0.0",
		Description: "演示固定列表与CRUD功能",
	},
	Menus: []plugins.Menu{
		{
			Key:         "plugin_demo",
			Title:       "演示插件",
			Icon:        "fa-solid fa-puzzle-piece",
			EventType:   "custom",
			CustomEvent: `() => loadJsonPage("/api/plugins/demo/page")`,
			Order:       100,
			Permission:  "demo.view",
		},
	},
	Permissions: []plugins.Permission{
		{Name: "demo.view", Title: "查看演示列表"},
		{Name: "demo.edit", Title: "编辑演示列表"},
	},
	APIRoutes: []plugins.APIRoute{
		{Method: "GET", Path: "/api/plugins/demo/items", RequiredPermission: "demo.view"},
		{Method: "POST", Path: "/api/plugins/demo/items", RequiredPermission: "demo.edit"},
		{Method: "PUT", Path: "/api/plugins/demo/items/:id", RequiredPermission: "demo.edit"},
		{Method: "DELETE", Path: "/api/plugins/demo/items/:id", RequiredPermission: "demo.edit"},
		{Method: "GET", Path: "/api/plugins/demo/page", RequiredPermission: "demo.view"},
	},
	Frontend: []plugins.FrontendResource{
		{Name: "demo.page", Path: "modules/demo/frontend/page.json"},
	},
	Tables: []plugins.Table{
		{Name: "demo_items", Schema: "ID, Name, Description, CreatedAt, UpdatedAt, CreatedBy"},
	},
	Lifecycle: &DemoLifecycle{},
	Router:    backend.RegisterRoutes,
}

func Permissions() []plugins.Permission {
	return []plugins.Permission{
		{Name: "demo.view", Title: "查看演示列表"},
		{Name: "demo.edit", Title: "编辑演示列表"},
	}
}

func Menus() []plugins.Menu {
	return []plugins.Menu{
		{
			Key:         "plugin_demo",
			Title:       "演示插件",
			Icon:        "fa-solid fa-puzzle-piece",
			EventType:   "custom",
			CustomEvent: `() => loadJsonPage("/api/plugins/demo/page")`,
			Order:       100,
			Permission:  "demo.view",
		},
	}
}
