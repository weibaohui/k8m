package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/plugins/modules/inspection/controller"
	"k8s.io/klog/v2"
)

// RegisterPluginAdminRoutes 注册集群巡检插件的管理员路由 - Gin到Chi迁移
// 使用插件内部的 controller 包，完全自包含
func RegisterPluginAdminRoutes(arg chi.Router) {
	controller.RegisterAdminScheduleRoutes(arg)
	controller.RegisterAdminRecordRoutes(arg)
	controller.RegisterAdminLuaScriptRoutes(arg)
	klog.V(6).Infof("注册集群巡检插件管理路由(admin)")
}
