package route

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/demo/admin"
	"k8s.io/klog/v2"
)

// RegisterPluginAdminRoutes 注册Demo插件的插件管理员类（admin）路由
func RegisterPluginAdminRoutes(arg *gin.RouterGroup) {
	g := arg.Group("/plugins/" + modules.PluginNameDemo)
	// 列表
	g.GET("/items", admin.List)
	// 新增
	g.POST("/items", admin.Create)
	// 更新
	g.POST("/items/:id", admin.Update)
	// 删除
	g.POST("/remove/items/:id", admin.Delete)

	klog.V(6).Infof("注册demo插件管理路由(admin)")
}
