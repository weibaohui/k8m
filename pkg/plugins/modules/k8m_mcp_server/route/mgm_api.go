package route

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/plugins/modules/k8m_mcp_server/server"
	"k8s.io/klog/v2"
)

func RegisterRootRoutes(arg *gin.RouterGroup) {
	g := arg.Group("/mcp/k8m")

	sseServer := server.GetMcpSSEServer()

	g.GET("/sse", server.Adapt(sseServer.SSEHandler))
	g.POST("/sse", server.Adapt(sseServer.SSEHandler))
	g.POST("/message", server.Adapt(sseServer.MessageHandler))
	g.GET("/:key/sse", server.Adapt(sseServer.SSEHandler))
	g.POST("/:key/sse", server.Adapt(sseServer.SSEHandler))
	g.POST("/:key/message", server.Adapt(sseServer.MessageHandler))

	klog.V(6).Infof("注册k8m_mcp_server插件管理路由(mgm)")
}
