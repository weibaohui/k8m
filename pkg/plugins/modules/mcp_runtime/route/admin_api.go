package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/mcp_runtime/admin"
	"k8s.io/klog/v2"
)

func RegisterPluginAdminRoutes(arg chi.Router) {
	g := arg.Group("/plugins/" + modules.PluginNameMCPRuntime)

	serverCtrl := &admin.ServerController{}
	g.GET("/server/list", serverCtrl.List)
	g.GET("/server/connect/:name", serverCtrl.Connect)
	g.POST("/server/delete", serverCtrl.Delete)
	g.POST("/server/save", serverCtrl.Save)
	g.POST("/server/save/id/:id/status/:status", serverCtrl.QuickSave)
	g.GET("/server/log/list", serverCtrl.MCPLogList)

	toolCtrl := &admin.ToolController{}
	g.GET("/tool/server/:name/list", toolCtrl.List)
	g.POST("/tool/save/id/:id/status/:status", toolCtrl.QuickSave)

	klog.V(6).Infof("注册 MCP 插件管理路由(admin)")
}

func RegisterPluginMgmRoutes(arg chi.Router) {
	mgm := arg.Group("/plugins/" + modules.PluginNameMCPRuntime)

	keyCtrl := &admin.KeyController{}
	mgm.GET("/user/profile/mcp_keys/list", keyCtrl.List)
	mgm.POST("/user/profile/mcp_keys/create", keyCtrl.Create)
	mgm.POST("/user/profile/mcp_keys/delete/:id", keyCtrl.Delete)

	klog.V(6).Infof("注册 MCP 插件管理路由(mgm)")
}
