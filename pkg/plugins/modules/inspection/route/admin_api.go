package route

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/plugins/modules/inspection/controller"
	"k8s.io/klog/v2"
)

// RegisterPluginAdminRoutes 注册集群巡检插件的管理员路由
// 使用插件内部的 controller 包，完全自包含
func RegisterPluginAdminRoutes(arg *gin.RouterGroup) {
	controller.RegisterAdminScheduleRoutes(arg)
	controller.RegisterAdminRecordRoutes(arg)
	controller.RegisterAdminLuaScriptRoutes(arg)
	klog.V(6).Infof("注册集群巡检插件管理路由(admin)")
}
