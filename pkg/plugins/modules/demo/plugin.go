package demo

import "github.com/weibaohui/k8m/pkg/plugins"

var ModuleDef = plugins.Module{
	Meta: plugins.Meta{
		Name:        "demo",
		Title:       "演示插件",
		Version:     "1.0.0",
		Description: "演示固定列表与CRUD功能",
	},
	Menus: []plugins.Menu{
		{ID: "demo.list", Title: "演示列表", Path: "/api/plugins/demo/page", Permission: "demo.view"},
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
}
