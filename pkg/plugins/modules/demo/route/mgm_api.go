package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/demo/mgm"
	"github.com/weibaohui/k8m/pkg/response"
	"k8s.io/klog/v2"
)

// RegisterMgmRoutes 注册Demo插件的管理类（mgm）路由
// 从 gin 切换到 chi，使用直接路由方法替代 gin.Group
func RegisterManagementRoutes(mrg chi.Router) {
	prefix := "/plugins/" + modules.PluginNameDemo
	mrg.Get(prefix+"/items", response.Adapter(mgm.List))
	mrg.Post(prefix+"/items", response.Adapter(mgm.Create))
	mrg.Post(prefix+"/items/{id}", response.Adapter(mgm.Update))
	mrg.Post(prefix+"/remove/items/{id}", response.Adapter(mgm.Delete))

	klog.V(6).Infof("注册demo插件管理路由")
}
