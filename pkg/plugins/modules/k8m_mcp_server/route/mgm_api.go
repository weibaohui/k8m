package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/plugins/modules/k8m_mcp_server/server"
	"github.com/weibaohui/k8m/pkg/response"
	"k8s.io/klog/v2"
)

// RegisterRootRoutes 注册路由
// 从 gin 切换到 chi，使用 chi.Router 替代 gin.RouterGroup
func RegisterRootRoutes(r chi.Router) {
	sseServer := server.GetMcpSSEServer()

	r.Get("/mcp/k8m/sse", response.Adapter(server.Adapt(sseServer.SSEHandler)))
	r.Post("/mcp/k8m/sse", response.Adapter(server.Adapt(sseServer.SSEHandler)))
	r.Post("/mcp/k8m/message", response.Adapter(server.Adapt(sseServer.MessageHandler)))
	r.Get("/mcp/k8m/{key}/sse", response.Adapter(server.Adapt(sseServer.SSEHandler)))
	r.Post("/mcp/k8m/{key}/sse", response.Adapter(server.Adapt(sseServer.SSEHandler)))
	r.Post("/mcp/k8m/{key}/message", response.Adapter(server.Adapt(sseServer.MessageHandler)))

	klog.V(6).Infof("注册k8m_mcp_server插件管理路由(mgm)")
}
