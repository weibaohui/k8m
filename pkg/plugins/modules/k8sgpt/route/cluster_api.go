package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/k8sgpt/admin"
	"github.com/weibaohui/k8m/pkg/response"
	"k8s.io/klog/v2"
)

func RegisterClusterRoutes(r chi.Router) {
	ctrl := &admin.Controller{}
	prefix := "/plugins/" + modules.PluginNameK8sGPT
	r.Get(prefix+"/kind/{kind}/run", response.Adapter(ctrl.ResourceRunAnalysis))

	klog.V(6).Infof("注册k8sgpt插件集群路由(cluster)")
}
