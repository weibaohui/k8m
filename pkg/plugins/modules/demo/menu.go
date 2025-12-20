package demo

import "github.com/weibaohui/k8m/pkg/plugins"

func Menus() []plugins.Menu {
	return []plugins.Menu{
		{ID: "demo.list", Title: "演示列表", Path: "/api/plugins/demo/page", Permission: "demo.view"},
	}
}

