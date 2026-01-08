package route

import (
	"github.com/go-chi/chi/v5"
	"k8s.io/klog/v2"
)

// RegisterClusterRoutes 注册Istio插件的集群相关路由
func RegisterClusterRoutes(crg chi.Router) {
	klog.V(6).Infof("注册istio插件路由(cluster)")
}
