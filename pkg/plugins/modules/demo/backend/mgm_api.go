package backend

import (
	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"
)

// RegisterMgmRoutes 注册Demo插件的管理类（mgm）路由
func RegisterMgmRoutes(mgm *gin.RouterGroup) {
	g := mgm.Group("/plugins/demo")
	// 列表
	g.GET("/items", List)
	// 新增
	g.POST("/items", Create)
	// 更新
	g.POST("/items/:id", Update)
	// 删除
	g.POST("/remove/items/:id", Delete)

	klog.V(6).Infof("注册插件管理路由(mgm): %s", "/items/:id")
}
