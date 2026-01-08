package route

import (
	"github.com/go-chi/chi/v5"
	"k8s.io/klog/v2"
)

// RegisterClusterRoutes 注册OpenKruise插件的集群相关路由
func RegisterClusterRoutes(crg chi.Router) {
	klog.V(6).Infof("注册openkruise插件路由(cluster)")
}
