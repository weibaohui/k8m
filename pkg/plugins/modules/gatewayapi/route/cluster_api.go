package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/gatewayapi/cluster"
	"github.com/weibaohui/k8m/pkg/response"
	"k8s.io/klog/v2"
)

// RegisterClusterRoutes 注册GatewayAPI插件的集群相关路由
func RegisterClusterRoutes(crg chi.Router) {
	prefix := "/plugins/" + modules.PluginNameGatewayAPI
	crg.Get(prefix+"/gateway_class/option_list", response.Adapter(cluster.GatewayClassOptionList))
	klog.V(6).Infof("注册gatewayapi插件路由(cluster)")
}
