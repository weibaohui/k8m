package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/mcp_runtime/admin"
	"github.com/weibaohui/k8m/pkg/response"
	"k8s.io/klog/v2"
)

// RegisterPluginAdminRoutes 注册插件管理路由

// Chi 中使用 chi.NewRouter() 创建子路由
func RegisterPluginAdminRoutes(arg chi.Router) {
	prefix := "/plugins/" + modules.PluginNameMCPRuntime
	serverCtrl := &admin.ServerController{}
	arg.Get(prefix+"/server/list", response.Adapter(serverCtrl.List))
	arg.Get(prefix+"/server/connect/{name}", response.Adapter(serverCtrl.Connect))
	arg.Post(prefix+"/server/delete", response.Adapter(serverCtrl.Delete))
	arg.Post(prefix+"/server/save", response.Adapter(serverCtrl.Save))
	arg.Post(prefix+"/server/save/id/{id}/status/{status}", response.Adapter(serverCtrl.QuickSave))
	arg.Get(prefix+"/server/log/list", response.Adapter(serverCtrl.MCPLogList))

	toolCtrl := &admin.ToolController{}
	arg.Get(prefix+"/tool/server/{name}/list", response.Adapter(toolCtrl.List))
	arg.Post(prefix+"/tool/save/id/{id}/status/{status}", response.Adapter(toolCtrl.QuickSave))

	klog.V(6).Infof("注册 MCP 插件管理路由(admin)")
}

// RegisterPluginMgmRoutes 注册插件管理路由

// Chi 中使用 chi.NewRouter() 创建子路由
func RegisterPluginMgmRoutes(arg chi.Router) {
	prefix := "/plugins/" + modules.PluginNameMCPRuntime
	keyCtrl := &admin.KeyController{}
	arg.Get(prefix+"/user/profile/mcp_keys/list", response.Adapter(keyCtrl.List))
	arg.Post(prefix+"/user/profile/mcp_keys/create", response.Adapter(keyCtrl.Create))
	arg.Post(prefix+"/user/profile/mcp_keys/delete/{id}", response.Adapter(keyCtrl.Delete))

	klog.V(6).Infof("注册 MCP 插件管理路由(mgm)")
}
