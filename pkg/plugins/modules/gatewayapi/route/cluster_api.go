package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/controller/gatewayapi"
	"k8s.io/klog/v2"
)

// RegisterClusterRoutes 注册GatewayAPI插件的集群相关路由
func RegisterClusterRoutes(crg chi.Router) {
	gatewayapi.RegisterRoutes(crg)
	klog.V(6).Infof("注册gatewayapi插件路由(cluster)")
}
