package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	mcp2 "github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/flag"
	"github.com/weibaohui/k8m/pkg/models"
	mcpModels "github.com/weibaohui/k8m/pkg/plugins/modules/mcp_runtime/models"
	"github.com/weibaohui/k8m/pkg/plugins/modules/mcp_runtime/service"
	"github.com/weibaohui/kom/mcp"
	"github.com/weibaohui/kom/mcp/tools"
	"k8s.io/klog/v2"
)

func Adapt(fn func() http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		handler := fn()
		handler.ServeHTTP(c.Writer, c.Request)
	}
}

func extractKey(path string) string {
	parts := strings.Split(path, "/")
	endpoints := []string{"sse", "message"}
	if len(parts) >= 5 && parts[1] == "mcp" && parts[2] == "k8m" && slice.Contain(endpoints, parts[4]) {
		return parts[3]
	}
	return ""
}

// createServerConfig 返回一个配置了 JWT 用户名提取、工具调用日志钩子及相关服务器和 SSE 选项的 MCP 服务器配置。
// 支持两种传递认证的方式，一是将mcpKey作为路径参数传递，二是将JWT token作为Authorization头部传递。
// mcpKey 与用户的信息绑定，mcpKey代表了一个用户
// token 与用户的JWT token绑定，代表了用户的权限，这个token与前端页面使用的jwt token 一致。
func createServerConfig(basePath string) *mcp.ServerConfig {
	cfg := flag.Init()

	var ctxFn = func(ctx context.Context, r *http.Request) context.Context {
		newCtx := context.Background()

		mcpKey := extractKey(r.URL.Path)
		if mcpKey != "" {
			username, err := service.McpService().GetUserByMCPKey(mcpKey)
			if err != nil {
				klog.V(6).Infof("Failed to extract username from mcpKey: %v", err)
			}
			if username != "" {
				newCtx = context.WithValue(newCtx, constants.JwtUserName, username)
				return newCtx
			}

		}

		auth := r.Header.Get("Authorization")
		if after, ok := strings.CutPrefix(auth, "Bearer "); ok {
			auth = after
		}
		klog.V(6).Infof("Authorization: %v", auth)
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
			resultInfo := mcpModels.MCPToolCallResult{
				ToolName:   toolName,
				Parameters: parameters,
				Result:     errStr,
				Error:      errStr,
			}
			host.LogToolExecution(ctx, toolName, serverName, parameters, resultInfo, 1)
		}
	}

	var actFn = func(ctx context.Context, id any, request *mcp2.CallToolRequest, result *mcp2.CallToolResult) {
		klog.V(8).Infof("CallToolRequest: %v", utils.ToJSON(request))
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

		resultInfo := mcpModels.MCPToolCallResult{
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
			server.WithDynamicBasePath(func(r *http.Request, sessionID string) string {
				key := extractKey(r.URL.Path)
				if key == "" {
					return basePath
				}
				return basePath + "/" + key
			}),
			server.WithStaticBasePath(basePath),
			server.WithSSEContextFunc(ctxFn),
		},
		AuthKey: constants.JwtUserName,
	}
}

func SaveYamlTemplateTool() mcp2.Tool {
	return mcp2.NewTool(
		"save_k8s_yaml_template",
		mcp2.WithDescription("保存Yaml为模版"),
		mcp2.WithString("cluster", mcp2.Description("模板适配集群（可为空）")),
		mcp2.WithString("yaml", mcp2.Required(), mcp2.Description("yaml模板内容，文本类型")),
		mcp2.WithString("name", mcp2.Description("模板名称")),
	)
}

func SaveYamlTemplateToolHandler(ctx context.Context, request mcp2.CallToolRequest) (*mcp2.CallToolResult, error) {
	username, ok := ctx.Value(constants.JwtUserName).(string)
	if !ok {
		username = ""
	}

	cluster := request.GetString("cluster", "")
	yaml := request.GetString("yaml", "")
	name := request.GetString("name", "")

	ct := models.CustomTemplate{
		Name:      name,
		Content:   yaml,
		Kind:      "",
		Cluster:   cluster,
		IsGlobal:  false,
		CreatedBy: username,
	}
	err := ct.Save(nil)
	if err != nil {
		return nil, err
	}
	return tools.TextResult("保存成功", nil)
}

func GetMcpSSEServer() *server.SSEServer {
	sc := createServerConfig("/mcp/k8m")
	serv := mcp.GetMCPServerWithOption(sc)
	serv.AddTool(SaveYamlTemplateTool(), SaveYamlTemplateToolHandler)
	return mcp.GetMCPSSEServerWithServerAndOption(serv, sc)
}
