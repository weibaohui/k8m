package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/plugins/modules/kubeconfig_export/mgm"
	"github.com/weibaohui/k8m/pkg/response"
	"k8s.io/klog/v2"
)

// RegisterManagementRoutes 注册Kubeconfig导出插件的管理类（mgm）路由
func RegisterManagementRoutes(mrg chi.Router) {
	prefix := "/plugins/kubeconfig_export"
	// 获取 kubeconfig 模板列表
	mrg.Get(prefix+"/templates", response.Adapter(mgm.ListTemplates))
	// 获取集群的 kubeconfig
	mrg.Get(prefix+"/cluster/{clusterID}/kubeconfig", response.Adapter(mgm.GetClusterKubeconfig))
	// 根据 ID 获取 kubeconfig
	mrg.Get(prefix+"/kubeconfig/{id}", response.Adapter(mgm.GetKubeConfigByID))
	// 导出 kubeconfig（根据 ID）- 改为 GET 请求
	mrg.Get(prefix+"/kubeconfig/{id}/export", response.Adapter(mgm.ExportKubeConfig))

	klog.V(6).Infof("注册kubeconfig_export插件管理路由")
}

// RegisterPluginAdminRoutes 注册Kubeconfig导出插件的插件管理员类路由
func RegisterPluginAdminRoutes(admin chi.Router) {
	// 前缀定义但未使用，保留用于未来的扩展
	// prefix := "/plugins/" + modules.PluginNameKubeconfigExport

	// 未来可以添加插件管理员相关路由

	klog.V(6).Infof("注册kubeconfig_export插件管理员路由")
}