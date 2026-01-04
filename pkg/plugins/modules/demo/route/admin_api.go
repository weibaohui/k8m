package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/demo/admin"
	"github.com/weibaohui/k8m/pkg/response"
	"k8s.io/klog/v2"
)

// RegisterPluginAdminRoutes 注册Demo插件的插件管理员类（admin）路由
func RegisterPluginAdminRoutes(arg chi.Router) {
	g := arg.Group("/plugins/" + modules.PluginNameDemo)
	g.Get("/items", response.Adapter(admin.List))
	g.Post("/items", response.Adapter(admin.Create))
	g.Post("/items/{id}", response.Adapter(admin.Update))
	g.Post("/remove/items/{id}", response.Adapter(admin.Delete))

	klog.V(6).Infof("注册demo插件管理路由(admin)")
}
