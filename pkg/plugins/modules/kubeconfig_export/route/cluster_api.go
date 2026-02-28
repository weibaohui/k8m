package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/kubeconfig_export/cluster"
	"github.com/weibaohui/k8m/pkg/response"
	"k8s.io/klog/v2"
)

// RegisterClusterRoutes 注册Kubeconfig导出插件的集群相关路由
func RegisterClusterRoutes(crg chi.Router) {
	prefix := "/plugins/" + modules.PluginNameKubeconfigExport
	// 生成 kubeconfig
	crg.Post(prefix+"/generate", response.Adapter(cluster.Generate))
	// 导出 kubeconfig
	crg.Get(prefix+"/export", response.Adapter(cluster.Export))

	klog.V(6).Infof("注册kubeconfig_export插件路由(cluster)")
}