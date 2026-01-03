package route

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/openapi/admin"
	"k8s.io/klog/v2"
)

func RegisterPluginMgmRoutes(arg *gin.RouterGroup) {
	g := arg.Group("/plugins/" + modules.PluginNameOpenAPI)
	ctrl := &admin.Controller{}

	g.GET("/api_keys/list", ctrl.List)
	g.POST("/api_keys/create", ctrl.Create)
	g.POST("/api_keys/delete/:id", ctrl.Delete)

	klog.V(6).Infof("注册openapi插件管理路由(mgm)")
}
