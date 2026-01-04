package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/mcp_runtime/admin"
	"github.com/weibaohui/k8m/pkg/response"
	"k8s.io/klog/v2"
)

// RegisterPluginAdminRoutes 注册插件管理路由
// 从 gin 切换到 chi，使用 chi.Router 替代 gin.RouterGroup
// Chi 中使用 chi.NewRouter() 创建子路由
func RegisterPluginAdminRoutes(arg chi.Router) {
	g := chi.NewRouter()

	serverCtrl := &admin.ServerController{}
	g.Get("/server/list", response.Adapter(serverCtrl.List))
	g.Get("/server/connect/{name}", response.Adapter(serverCtrl.Connect))
	g.Post("/server/delete", response.Adapter(serverCtrl.Delete))
	g.Post("/server/save", response.Adapter(serverCtrl.Save))
	g.Post("/server/save/id/{id}/status/{status}", response.Adapter(serverCtrl.QuickSave))
	g.Get("/server/log/list", response.Adapter(serverCtrl.MCPLogList))

	toolCtrl := &admin.ToolController{}
	g.Get("/tool/server/{name}/list", response.Adapter(toolCtrl.List))
	g.Post("/tool/save/id/{id}/status/{status}", response.Adapter(toolCtrl.QuickSave))

	arg.Mount("/plugins/"+modules.PluginNameMCPRuntime, g)

	klog.V(6).Infof("注册 MCP 插件管理路由(admin)")
}

// RegisterPluginMgmRoutes 注册插件管理路由
// 从 gin 切换到 chi，使用 chi.Router 替代 gin.RouterGroup
// Chi 中使用 chi.NewRouter() 创建子路由
func RegisterPluginMgmRoutes(arg chi.Router) {
	mgm := chi.NewRouter()

	keyCtrl := &admin.KeyController{}
	mgm.Get("/user/profile/mcp_keys/list", response.Adapter(keyCtrl.List))
	mgm.Post("/user/profile/mcp_keys/create", response.Adapter(keyCtrl.Create))
	mgm.Post("/user/profile/mcp_keys/delete/{id}", response.Adapter(keyCtrl.Delete))

	arg.Mount("/plugins/"+modules.PluginNameMCPRuntime, mgm)

	klog.V(6).Infof("注册 MCP 插件管理路由(mgm)")
}
