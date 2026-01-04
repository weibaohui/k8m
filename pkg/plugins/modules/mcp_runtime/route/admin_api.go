package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/mcp_runtime/admin"
	"github.com/weibaohui/k8m/pkg/response"
	"k8s.io/klog/v2"
)

func RegisterPluginAdminRoutes(arg chi.Router) {
	g := arg.Group("/plugins/" + modules.PluginNameMCPRuntime)

	serverCtrl := &admin.ServerController{}
	g.Get("/server/list", response.Adapter(serverCtrl.List))
	g.Get("/server/connect/{name}", response.Adapter(serverCtrl.Connect))
	g.Post("/server/delete", response.Adapter(serverCtrl.Delete))
	g.Post("/server/save", response.Adapter(serverCtrl.Save))
	g.Post("/server/save/id/{id}/status/{status}", response.Adapter(serverCtrl.QuickSave))
	g.Get("/server/log/list", response.Adapter(serverCtrl.MCPLogList))

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
