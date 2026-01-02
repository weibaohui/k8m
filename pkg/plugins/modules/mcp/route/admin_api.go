package route

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"k8s.io/klog/v2"
)

func RegisterPluginAdminRoutes(arg *gin.RouterGroup) {
	g := arg.Group("/plugins/" + modules.PluginNameMCP)

	serverCtrl := &admin.ServerController{}
	g.GET("/server/list", serverCtrl.List)
	g.GET("/server/connect/:name", serverCtrl.Connect)
	g.POST("/server/delete", serverCtrl.Delete)
	g.POST("/server/save", serverCtrl.Save)
	g.GET("/server/get/:id", serverCtrl.Get)

	toolCtrl := &admin.ToolController{}
	g.GET("/tool/list", toolCtrl.List)
	g.GET("/tool/server/:name/list", toolCtrl.ListByServer)
	g.POST("/tool/save", toolCtrl.Save)
	g.POST("/tool/delete", toolCtrl.Delete)
	g.POST("/tool/save/id/:id/status/:status", toolCtrl.QuickSave)

	logCtrl := &admin.LogController{}
	g.GET("/log/list", logCtrl.List)
	g.GET("/log/clear", logCtrl.Clear)

	klog.V(6).Infof("注册 MCP 插件管理路由(admin)")
}

func RegisterPluginMgmRoutes(arg *gin.RouterGroup) {
	mgm := arg.Group("/plugins/" + modules.PluginNameMCP)

	keyCtrl := &admin.KeyController{}
	mgm.GET("/keys/list", keyCtrl.List)
	mgm.POST("/keys/save", keyCtrl.Save)
	mgm.POST("/keys/delete", keyCtrl.Delete)
	mgm.GET("/keys/my/list", keyCtrl.MyList)
	mgm.GET("/keys/my/gen", keyCtrl.GenKey)
	mgm.GET("/keys/refresh/jwt", keyCtrl.RefreshJWT)

	klog.V(6).Infof("注册 MCP 插件管理路由(mgm)")
}
