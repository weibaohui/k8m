package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	mcp2 "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/flag"
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/mcp"
	"github.com/weibaohui/kom/mcp/metadata"
)

// createServerConfig 创建MCP服务器配置
func createServerConfig(basePath string) *metadata.ServerConfig {
	cfg := flag.Init()

	var ctxFn = func(ctx context.Context, r *http.Request) context.Context {
		auth := r.Header.Get("Authorization")
		newCtx := context.Background()
		if username, err := utils.GetUsernameFromToken(auth, cfg.JwtTokenSecret); err == nil {
			newCtx = context.WithValue(newCtx, constants.JwtUserName, username)
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
			resultInfo := service.MCPToolCallResult{
				ToolName:   toolName,
				Parameters: parameters,
				Result:     errStr,
				Error:      errStr,
			}
			host.LogToolExecution(ctx, toolName, serverName, parameters, resultInfo, 1)
		}
	}

	var actFn = func(ctx context.Context, id any, request *mcp2.CallToolRequest, result *mcp2.CallToolResult) {
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

		resultInfo := service.MCPToolCallResult{
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

	return &metadata.ServerConfig{
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

// GetMcpSSEServer 获取MCP SSE服务器
func GetMcpSSEServer(basePath string) *server.SSEServer {
	sc := createServerConfig(basePath)
	return mcp.GetMCPSSEServerWithOption(sc)
}
func adapt(fn func() http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		handler := fn()
		handler.ServeHTTP(c.Writer, c.Request)
	}
}
