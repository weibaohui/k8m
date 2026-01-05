package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/plugins/modules/k8sgpt/admin"
	"github.com/weibaohui/k8m/pkg/response"
	"k8s.io/klog/v2"
)

func RegisterClusterRoutes(r chi.Router) {
	ctrl := &admin.Controller{}

	r.Get("/k8s_gpt/kind/{kind}/run", response.Adapter(ctrl.ResourceRunAnalysis))
	r.Post("/k8s_gpt/cluster/{user_cluster}/run", response.Adapter(ctrl.ClusterRunAnalysis))
	r.Get("/k8s_gpt/cluster/{user_cluster}/result", response.Adapter(ctrl.GetClusterRunAnalysisResult))
	r.Get("/k8s_gpt/var", response.Adapter(ctrl.GetFields))

	klog.V(6).Infof("注册k8sgpt插件集群路由(cluster)")
}
