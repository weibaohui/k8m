package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/demo/mgm"
	"github.com/weibaohui/k8m/pkg/response"
	"k8s.io/klog/v2"
)

// RegisterMgmRoutes 注册Demo插件的管理类（mgm）路由
func RegisterManagementRoutes(mrg chi.Router) {
	g := mrg.Group("/plugins/" + modules.PluginNameDemo)

	g.Get("/items", response.Adapter(mgm.List))
	g.Post("/items", response.Adapter(mgm.Create))
	g.Post("/items/{id}", response.Adapter(mgm.Update))
	g.Post("/remove/items/{id}", response.Adapter(mgm.Delete))

	klog.V(6).Infof("注册demo插件管理路由")
}
