package param

import "github.com/gin-gonic/gin"

type Controller struct {
}

func RegisterParamRoutes(params *gin.RouterGroup) {
	ctrl := &Controller{}
	// 获取当前登录用户的角色，登录即可
	params.GET("/user/role", ctrl.UserRole)
	// 获取某个配置项
	params.GET("/config/:key", ctrl.Config)
	// 获取当前登录用户的集群列表,下拉列表
	params.GET("/cluster/option_list", ctrl.ClusterOptionList)
	// 获取当前登录用户的集群列表,table列表
	params.GET("/cluster/all", ctrl.ClusterTableList)
	// 获取当前软件版本信息
	params.GET("/version", ctrl.Version)
	// 获取helm 仓库列表
	params.GET("/helm/repo/option_list", ctrl.HelmRepoOptionList)
	// 获取翻转显示的指标列表
	params.GET("/condition/reverse/list", ctrl.Conditions)
}
