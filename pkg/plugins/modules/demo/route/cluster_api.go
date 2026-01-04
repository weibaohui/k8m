package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/demo/cluster"
	"github.com/weibaohui/k8m/pkg/response"
	"k8s.io/klog/v2"
)

// RegisterClusterRoutes 注册Demo插件的集群相关路由
// 从 gin 切换到 chi，使用直接路由方法替代 gin.Group
func RegisterClusterRoutes(crg chi.Router) {
	prefix := "/plugins/" + modules.PluginNameDemo
	crg.Get(prefix+"/items", response.Adapter(cluster.List))
	crg.Post(prefix+"/items", response.Adapter(cluster.Create))
	crg.Post(prefix+"/items/{id}", response.Adapter(cluster.Update))
	crg.Post(prefix+"/remove/items/{id}", response.Adapter(cluster.Delete))

	klog.V(6).Infof("注册demo插件路由(cluster)")
}
