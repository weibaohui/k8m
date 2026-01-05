package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/webhook/admin"
	"github.com/weibaohui/k8m/pkg/response"
	"k8s.io/klog/v2"
)

// RegisterPluginAdminRoutes 注册 webhook 插件管理路由

func RegisterPluginAdminRoutes(r chi.Router) {
	ctrl := &admin.Controller{}

	r.Get("/plugins/"+modules.PluginNameWebhook+"/list", response.Adapter(ctrl.WebhookList))
	r.Post("/plugins/"+modules.PluginNameWebhook+"/delete/{ids}", response.Adapter(ctrl.WebhookDelete))
	r.Post("/plugins/"+modules.PluginNameWebhook+"/save", response.Adapter(ctrl.WebhookSave))
	r.Post("/plugins/"+modules.PluginNameWebhook+"/id/{id}/test", response.Adapter(ctrl.WebhookTest))
	r.Get("/plugins/"+modules.PluginNameWebhook+"/option_list", response.Adapter(ctrl.WebhookOptionList))

	r.Get("/plugins/"+modules.PluginNameWebhook+"/records", response.Adapter(ctrl.WebhookRecordList))
	r.Get("/plugins/"+modules.PluginNameWebhook+"/records/{id}", response.Adapter(ctrl.WebhookRecordDetail))
	r.Get("/plugins/"+modules.PluginNameWebhook+"/records/statistics", response.Adapter(ctrl.WebhookRecordStatistics))

	klog.V(6).Infof("注册webhook插件管理路由(admin)")
}
