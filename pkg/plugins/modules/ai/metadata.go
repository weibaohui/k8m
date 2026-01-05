package ai

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/ai/route"
)

var Metadata = plugins.Module{
	Meta: plugins.Meta{
		Name:        modules.PluginNameAI,
		Title:       "AI 插件",
		Version:     "1.0.0",
		Description: "AI功能插件，提供K8s资源智能分析、事件问诊、日志分析、Cron表达式解析等功能。支持自定义AI模型配置。",
	},
	Tables: []string{
		"ai_models",
		"ai_prompts",
	},
	Menus: []plugins.Menu{
		{
			Key:   "plugin_ai_index",
			Title: "AI 管理",
			Icon:  "fa-solid fa-brain",
			Order: 30,
			Children: []plugins.Menu{
				{
					Key:         "plugin_ai_model",
					Title:       "AI模型配置",
					Icon:        "fa-solid fa-robot",
					Show:        "isPlatformAdmin()==true",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/admin/config/ai_model_config")`,
					Order:       10,
				},
				{
					Key:         "plugin_ai_prompt",
					Title:       "Prompt模板管理",
					Icon:        "fa-solid fa-wand-magic-sparkles",
					Show:        "isPlatformAdmin()==true",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/admin/config/ai_prompt_config")`,
					Order:       20,
				},
			},
		},
	},
	Dependencies:     []string{},
	RunAfter:         []string{},
	Lifecycle:        &AILifecycle{},
	ManagementRouter: route.RegisterManagementRoutes,
}
