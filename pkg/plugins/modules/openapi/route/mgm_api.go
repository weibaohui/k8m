package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/openapi/admin"
	"github.com/weibaohui/k8m/pkg/response"
	"k8s.io/klog/v2"
)

// RegisterPluginMgmRoutes 注册插件管理路由

// Chi 中使用 chi.NewRouter() 创建子路由
func RegisterPluginMgmRoutes(arg chi.Router) {
	g := chi.NewRouter()
	ctrl := &admin.Controller{}

	g.Get("/api_keys/list", response.Adapter(ctrl.List))
	g.Post("/api_keys/create", response.Adapter(ctrl.Create))
	g.Post("/api_keys/delete/{id}", response.Adapter(ctrl.Delete))

	arg.Mount("/plugins/"+modules.PluginNameOpenAPI, g)

	klog.V(6).Infof("注册openapi插件管理路由(mgm)")
}
