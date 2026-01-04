package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/demo/admin"
	"github.com/weibaohui/k8m/pkg/response"
	"k8s.io/klog/v2"
)

// RegisterPluginAdminRoutes 注册Demo插件的插件管理员类（admin）路由
// 从 gin 切换到 chi，使用 chi.RoutePrefix 替代 gin.Group
func RegisterPluginAdminRoutes(arg chi.Router) {
	prefix := "/plugins/" + modules.PluginNameDemo
	arg.Get(prefix+"/items", response.Adapter(admin.List))
	arg.Post(prefix+"/items", response.Adapter(admin.Create))
	arg.Post(prefix+"/items/{id}", response.Adapter(admin.Update))
	arg.Post(prefix+"/remove/items/{id}", response.Adapter(admin.Delete))

	klog.V(6).Infof("注册demo插件管理路由(admin)")
}
