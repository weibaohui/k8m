package yaml_editor

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/yaml_editor/route"
)

var Metadata = plugins.Module{
	Meta: plugins.Meta{
		Name:        modules.PluginNameYamlEditor,
		Title:       "YAML 编辑器",
		Version:     "1.0.0",
		Description: "YAML 编辑器，支持 YAML 配置的应用、删除、模板管理和历史记录功能。",
	},
	Tables: []string{
		"yaml_editor_templates",
	},
	Menus: []plugins.Menu{
		{
			Key:   "plugin_yaml_editor_index",
			Title: "YAML 编辑器",
			Icon:  "fa-solid fa-code",
			Order: 2,
			Children: []plugins.Menu{
				{
					Key:         "plugin_yaml_editor_main",
					Title:       "YAML管理",
					Icon:        "fa-solid fa-file-code",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/yaml_editor/main")`,
					Order:       10,
				},
			},
		},
	},
	Dependencies:     []string{},
	RunAfter:         []string{},
	Lifecycle:        &YamlEditorLifecycle{},
	ClusterRouter:    route.RegisterClusterRoutes,
	ManagementRouter: route.RegisterManagementRoutes,
}
