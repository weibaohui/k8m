package demo

import "github.com/weibaohui/k8m/pkg/plugins"

func Permissions() []plugins.Permission {
	return []plugins.Permission{
		{Name: "demo.view", Title: "查看演示列表"},
		{Name: "demo.edit", Title: "编辑演示列表"},
	}
}

