package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/k8sgpt/admin"
	"github.com/weibaohui/k8m/pkg/response"
	"k8s.io/klog/v2"
)

func RegisterMgmRoutes(r chi.Router) {
	ctrl := &admin.Controller{}
	prefix := "/plugins/" + modules.PluginNameK8sGPT
	r.Post(prefix+"/cluster/{cluster}/run", response.Adapter(ctrl.ClusterRunAnalysisMgm))
	r.Get(prefix+"/cluster/{cluster}/result", response.Adapter(ctrl.GetClusterRunAnalysisResultMgm))

	klog.V(6).Infof("注册k8sgpt插件管理路由(mgm)")
}
