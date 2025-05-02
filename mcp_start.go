package main

import (
	"context"
	"fmt"
	"net/http"

	mcp2 "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/constants"
	mcp3 "github.com/weibaohui/k8m/pkg/mcp"
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/mcp"
	"github.com/weibaohui/kom/mcp/metadata"
	"k8s.io/klog/v2"
)

func MCPStart(version string, port int) {
	var ctxFn = func(ctx context.Context, r *http.Request) context.Context {
		username := r.Header.Get(constants.JwtUserName)
		role := r.Header.Get(constants.JwtUserRole)
		channel := server.GetRouteParam(ctx, "channel")
		newCtx := context.Background()
		if channel == "inner" {
			// 发起mcp调用请求时注入用户名、角色信息
			newCtx = context.WithValue(ctx, constants.JwtUserName, username)
			ctx = context.WithValue(newCtx, constants.JwtUserRole, role)
			klog.V(6).Infof("mcp inner request, username: %s, role: %s", username, role)
		} else {
			if user, err := service.UserService().GetUserByMCPKey(channel); err == nil {
				newCtx = context.WithValue(ctx, constants.JwtUserName, user)
			}
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
			resultInfo := mcp3.ToolCallResult{
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
		
		resultInfo := mcp3.ToolCallResult{
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
	cfg := metadata.ServerConfig{
		Name:    "k8m mcp server",
		Version: version,
		Port:    port,
		ServerOptions: []server.ServerOption{
			server.WithResourceCapabilities(false, false),
			server.WithPromptCapabilities(false),
			server.WithLogging(),
			server.WithHooks(hooks),
		},
		SSEOption: []server.SSEOption{
			server.WithSSEPattern("/:channel/sse"),
			server.WithSSEContextFunc(ctxFn),
		},
		AuthKey:     constants.JwtUserName,
		AuthRoleKey: constants.JwtUserRole,
	}
	mcp.RunMCPServerWithOption(&cfg)
}
