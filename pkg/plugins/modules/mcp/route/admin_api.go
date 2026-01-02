package route

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/mcp/admin"
	"k8s.io/klog/v2"
)

func RegisterPluginAdminRoutes(arg *gin.RouterGroup) {
	g := arg.Group("/plugins/" + modules.PluginNameMCP)

	serverCtrl := &admin.ServerController{}
	g.GET("/server/list", serverCtrl.List)
	g.GET("/server/connect/:name", serverCtrl.Connect)
	g.POST("/server/delete", serverCtrl.Delete)
	g.POST("/server/save", serverCtrl.Save)
	g.POST("/server/save/id/:id/status/:status", serverCtrl.QuickSave)
	g.GET("/log/list", serverCtrl.MCPLogList)

	toolCtrl := &admin.ToolController{}
	g.GET("/tool/server/:name/list", toolCtrl.List)
	g.POST("/tool/save/id/:id/status/:status", toolCtrl.QuickSave)

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
