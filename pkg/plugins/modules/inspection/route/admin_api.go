package route

import (
	"github.com/gin-gonic/gin"
	adminInspection "github.com/weibaohui/k8m/pkg/controller/admin/inspection"
	"k8s.io/klog/v2"
)

// RegisterPluginAdminRoutes 注册集群巡检插件的管理员路由
// 这里直接复用现有的 admin/inspection 控制器，将其挂载到插件管理器提供的 /admin 分组下。
// 插件启用后，这些路由才会生效；关闭插件则不再注册巡检相关接口。
func RegisterPluginAdminRoutes(arg *gin.RouterGroup) {
	adminInspection.RegisterAdminScheduleRoutes(arg)
	adminInspection.RegisterAdminRecordRoutes(arg)
	adminInspection.RegisterAdminLuaScriptRoutes(arg)
	klog.V(6).Infof("注册集群巡检插件管理路由(admin)")
}
