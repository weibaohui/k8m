package param

import (
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/response"
)

type Controller struct {
}

func RegisterParamRoutes(r chi.Router) {
	ctrl := &Controller{}
	// 获取当前登录用户的角色，登录即可
	r.Get("/user/role", response.Adapter(ctrl.UserRole))
	// 获取某个配置项
	r.Get("/config/{key}", response.Adapter(ctrl.Config))
	// 获取当前登录用户的集群列表,下拉列表
	r.Get("/cluster/option_list", response.Adapter(ctrl.ClusterOptionList))
	// 获取当前登录用户的集群列表,table列表
	r.Get("/cluster/all", response.Adapter(ctrl.ClusterTableList))
	// 获取当前软件版本信息
	r.Get("/version", response.Adapter(ctrl.Version))
	// 获取helm 仓库列表
	r.Get("/helm/repo/option_list", response.Adapter(ctrl.HelmRepoOptionList))
	// 获取翻转显示的指标列表
	r.Get("/condition/reverse/list", response.Adapter(ctrl.Conditions))
}
