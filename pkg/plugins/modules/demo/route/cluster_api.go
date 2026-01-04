package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/demo/cluster"
	"k8s.io/klog/v2"
)

// RegisterClusterRoutes 注册Demo插件的集群相关路由
func RegisterClusterRoutes(crg chi.Router) {
	g := crg.Group("/plugins/" + modules.PluginNameDemo)

	// 列表
	g.GET("/items", cluster.List)
	// 新增
	g.POST("/items", cluster.Create)
	// 更新
	g.POST("/items/:id", cluster.Update)
	// 删除
	g.POST("/remove/items/:id", cluster.Delete)

	klog.V(6).Infof("注册demo插件路由(cluster)")
}
