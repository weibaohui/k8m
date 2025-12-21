package route

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/plugins/modules/demo/mgm"
	"k8s.io/klog/v2"
)

// RegisterMgmRoutes 注册Demo插件的管理类（mgm）路由
func RegisterManagementRoutes(mrg *gin.RouterGroup) {
	g := mrg.Group("/plugins/demo")
	// 列表
	g.GET("/items", mgm.List)
	// 新增
	g.POST("/items", mgm.Create)
	// 更新
	g.POST("/items/:id", mgm.Update)
	// 删除
	g.POST("/remove/items/:id", mgm.Delete)

	klog.V(6).Infof("注册demo插件管理路由")
}
