package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	mcp2 "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/flag"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/mcp"
	"k8s.io/klog/v2"
)

// createServerConfig 根据指定的 basePath 创建并返回 MCP 服务器的配置。
// 配置包括 JWT 用户名提取的上下文函数、工具调用错误和成功后的日志钩子、服务器选项及 SSE 相关设置。
func createServerConfig(basePath string) *mcp.ServerConfig {
	cfg := flag.Init()

	var ctxFn = func(ctx context.Context, r *http.Request) context.Context {
		auth := r.Header.Get("Authorization")
		// 处理 Bearer 前缀
		if strings.HasPrefix(auth, "Bearer ") {
			auth = strings.TrimPrefix(auth, "Bearer ")
		}
		klog.V(6).Infof("Authorization: %v", auth)
		newCtx := context.Background()
		if username, err := utils.GetUsernameFromToken(auth, cfg.JwtTokenSecret); err == nil {
			klog.V(6).Infof("Extracted username from token: %v", username)
			newCtx = context.WithValue(newCtx, constants.JwtUserName, username)
		} else {
			klog.V(6).Infof("Failed to extract username from token: %v", err)
		}
		return newCtx
	}

	var errFn = func(ctx context.Context, id any, method mcp2.MCPMethod, message any, err error) {
		if request, ok := message.(*mcp2.CallToolRequest); ok {
			errStr := fmt.Sprintf("%v", err)
			host := service.McpService().Host()
			toolName := request.Params.Name
			serverName := host.GetServerNameByToolName(toolName)
			parameters := request.Params.Arguments
			resultInfo := models.MCPToolCallResult{
				ToolName:   toolName,
				Parameters: parameters,
				Result:     errStr,
				Error:      errStr,
			}
			host.LogToolExecution(ctx, toolName, serverName, parameters, resultInfo, 1)
		}
	}

	var actFn = func(ctx context.Context, id any, request *mcp2.CallToolRequest, result *mcp2.CallToolResult) {
		// 记录工具调用请求
		klog.V(6).Infof("CallToolRequest: %v", utils.ToJSON(request))
		host := service.McpService().Host()
		toolName := request.Params.Name
		serverName := host.GetServerNameByToolName(toolName)
		parameters := request.Params.Arguments
		var resultStr string
		var errStr string
		resultStr = utils.ToJSON(result)
		if result.IsError {
			errStr = resultStr
		}

		resultInfo := models.MCPToolCallResult{
			ToolName:   toolName,
			Parameters: parameters,
			Result:     resultStr,
			Error:      errStr,
		}
		host.LogToolExecution(ctx, toolName, serverName, parameters, resultInfo, 1)
	}

	hooks := &server.Hooks{
		OnError:         []server.OnErrorHookFunc{errFn},
		OnAfterCallTool: []server.OnAfterCallToolFunc{actFn},
	}

	return &mcp.ServerConfig{
		Name:    "k8m mcp server",
		Version: cfg.Version,
		ServerOptions: []server.ServerOption{
			server.WithResourceCapabilities(false, false),
			server.WithPromptCapabilities(false),
			server.WithLogging(),
			server.WithHooks(hooks),
		},
		SSEOption: []server.SSEOption{
			server.WithBasePath(basePath),
			server.WithSSEContextFunc(ctxFn),
		},
		AuthKey: constants.JwtUserName,
	}
}

// GetMcpSSEServer 根据指定的基础路径创建并返回一个配置好的MCP SSE服务器实例。
func GetMcpSSEServer(basePath string) *server.SSEServer {
	sc := createServerConfig(basePath)
	return mcp.GetMCPSSEServerWithOption(sc)
}

// adapt 将标准的 http.Handler 适配为 Gin 框架可用的处理函数。
func adapt(fn func() http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		handler := fn()
		handler.ServeHTTP(c.Writer, c.Request)
	}
}
