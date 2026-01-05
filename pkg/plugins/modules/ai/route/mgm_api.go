package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/ai/controller"
	"github.com/weibaohui/k8m/pkg/response"
	"k8s.io/klog/v2"
)

func RegisterManagementRoutes(arg chi.Router) {
	prefix := "/plugins/" + modules.PluginNameAI

	ctrl := &controller.Controller{}

	arg.Get(prefix+"/chat/event", response.Adapter(ctrl.Event))
	arg.Get(prefix+"/chat/log", response.Adapter(ctrl.Log))
	arg.Get(prefix+"/chat/cron", response.Adapter(ctrl.Cron))
	arg.Get(prefix+"/chat/describe", response.Adapter(ctrl.Describe))
	arg.Get(prefix+"/chat/resource", response.Adapter(ctrl.Resource))
	arg.Get(prefix+"/chat/any_question", response.Adapter(ctrl.AnyQuestion))
	arg.Get(prefix+"/chat/any_selection", response.Adapter(ctrl.AnySelection))
	arg.Get(prefix+"/chat/example", response.Adapter(ctrl.Example))
	arg.Get(prefix+"/chat/example/field", response.Adapter(ctrl.FieldExample))
	arg.Get(prefix+"/chat/ws_chatgpt", response.Adapter(ctrl.GPTShell))
	arg.Get(prefix+"/chat/ws_chatgpt/history", response.Adapter(ctrl.History))
	arg.Get(prefix+"/chat/ws_chatgpt/history/reset", response.Adapter(ctrl.Reset))
	arg.Get(prefix+"/chat/k8s_gpt/resource", response.Adapter(ctrl.K8sGPTResource))

	klog.V(6).Infof("注册 AI 插件管理路由")
}

func RegisterPluginAdminRoutes(arg chi.Router) {
	prefix := "/plugins/" + modules.PluginNameAI

	ac := &controller.AdminAIPromptController{}
	arg.Get(prefix+"/ai_prompt/list", response.Adapter(ac.AIPromptList))
	arg.Post(prefix+"/ai_prompt/delete/{ids}", response.Adapter(ac.AIPromptDelete))
	arg.Post(prefix+"/ai_prompt/save", response.Adapter(ac.AIPromptSave))
	arg.Post(prefix+"/ai_prompt/load", response.Adapter(ac.AIPromptLoad))
	arg.Get(prefix+"/ai_prompt/option_list", response.Adapter(ac.AIPromptOptionList))
	arg.Get(prefix+"/ai_prompt/types", response.Adapter(ac.AIPromptTypes))
	arg.Post(prefix+"/ai_prompt/toggle/{id}", response.Adapter(ac.AIPromptToggle))                  // 添加启用/禁用路由
	arg.Post(prefix+"/ai_prompt/id/{id}/enabled/{enabled}", response.Adapter(ac.AIPromptQuickSave)) // 快捷保存启用状态

	klog.V(6).Infof("注册 AI 插件 提示词 管理路由")
}
