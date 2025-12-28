package route

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/eventhandler/admin"
	"k8s.io/klog/v2"
)

// RegisterPluginAdminRoutes 中文函数注释：注册事件转发插件的管理员路由（平台管理员）。
func RegisterPluginAdminRoutes(arg *gin.RouterGroup) {
	g := arg.Group("/plugins/" + modules.PluginNameEventHandler)
	ctrl := &admin.Controller{}

	g.GET("/setting/get", ctrl.GetSetting)
	g.POST("/setting/update", ctrl.UpdateSetting)

	g.GET("/list", ctrl.List)
	g.POST("/save", ctrl.Save)
	g.POST("/delete/:ids", ctrl.Delete)
	g.POST("/save/id/:id/status/:enabled", ctrl.QuickSave)

	klog.V(6).Infof("注册事件转发插件管理路由(admin)")
}
