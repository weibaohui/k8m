package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/heartbeat/controller"
	"github.com/weibaohui/k8m/pkg/response"
	"k8s.io/klog/v2"
)

// RegisterPluginAdminRoutes 注册插件管理员API路由
func RegisterPluginAdminRoutes(r chi.Router) {
	// 心跳配置相关API

	ctrl := controller.Controller{}
	r.Get("/plugins/"+modules.PluginNameHeartbeat+"/config", response.Adapter(ctrl.GetHeartbeatConfig))
	r.Post("/plugins/"+modules.PluginNameHeartbeat+"/config", response.Adapter(ctrl.SaveHeartbeatConfig))
	r.Get("/plugins/"+modules.PluginNameHeartbeat+"/status", response.Adapter(ctrl.GetHeartbeatStatus))

	klog.V(6).Infof("注册heartbeat插件管理路由(admin)")
}
