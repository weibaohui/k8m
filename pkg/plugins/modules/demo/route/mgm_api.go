package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/demo/mgm"
	"k8s.io/klog/v2"
)

// RegisterMgmRoutes 注册Demo插件的管理类（mgm）路由
func RegisterManagementRoutes(mrg chi.Router) {
	g := mrg.Group("/plugins/" + modules.PluginNameDemo)

	// 列表
	g.GET("/items", mgm.List)
	// 新增
	g.POST("/items", mgm.Create)
	// 更新
	g.POST("/items/:id", mgm.Update)
	// 删除
	g.POST("/remove/items/:id", mgm.Delete)

	klog.V(6).Infof("注册demo插件管理路由")
}
