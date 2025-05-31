package main

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
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/mcp"
	"github.com/weibaohui/kom/mcp/tools"
	"k8s.io/klog/v2"
)

// createServerConfig 根据指定的 basePath 创建并返回 MCP 服务器的配置。
// createServerConfig 返回一个配置了 JWT 用户名提取、工具调用日志钩子及相关服务器和 SSE 选项的 MCP 服务器配置。
// 配置包括从 HTTP Authorization 头部提取并解析 JWT 用户名的上下文函数，工具调用错误和成功后的日志记录钩子，以及基础路径和认证键等服务器参数。
func createServerConfig(basePath string) *mcp.ServerConfig {
	cfg := flag.Init()

	var ctxFn = func(ctx context.Context, r *http.Request) context.Context {
		newCtx := context.Background()

		mcpKey := extractKey(r.URL.Path)
		if mcpKey != "" {
			username, err := service.UserService().GetUserByMCPKey(mcpKey)
			if err != nil {
				klog.V(6).Infof("Failed to extract username from mcpKey: %v", err)
			}
			if username != "" {
				newCtx = context.WithValue(newCtx, constants.JwtUserName, username)
				return newCtx
			}

		}

		auth := r.Header.Get("Authorization")
		// 处理 Bearer 前缀
		if strings.HasPrefix(auth, "Bearer ") {
			auth = strings.TrimPrefix(auth, "Bearer ")
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

// SaveYamlTemplateTool 返回一个用于保存 Kubernetes YAML 模板的 MCP 工具定义。
func SaveYamlTemplateTool() mcp2.Tool {
	return mcp2.NewTool(
		"save_k8s_yaml_template",
		mcp2.WithDescription("保存Yaml为模版"),
		mcp2.WithString("cluster", mcp2.Description("模板适配集群（可为空）")),
		mcp2.WithString("yaml", mcp2.Required(), mcp2.Description("yaml模板内容，文本类型")),
		mcp2.WithString("name", mcp2.Description("模板名称")),
	)
}

// SaveYamlTemplateToolHandler 处理保存 Kubernetes YAML 模板的工具调用请求。
// 该函数从上下文中提取用户名，从请求参数中提取集群信息、YAML 内容和模板名称，
// 然后创建一个自定义模板对象并保存到存储中。如果保存成功，返回包含成功消息的调用结果；
// 若保存失败，则返回错误。
func SaveYamlTemplateToolHandler(ctx context.Context, request mcp2.CallToolRequest) (*mcp2.CallToolResult, error) {
	// 从上下文中提取 JWT 用户名
	username, ok := ctx.Value(constants.JwtUserName).(string)
	// 若提取失败，将用户名置为空字符串
	if !ok {
		username = ""
	}

	// 尝试从请求参数中提取集群信息
	cluster := request.GetString("cluster", "")
	// 尝试从请求参数中提取 YAML 内容
	yaml := request.GetString("yaml", "")
	// 尝试从请求参数中提取模板名称
	name := request.GetString("name", "")

	// 创建自定义模板对象
	ct := models.CustomTemplate{
		Name:      name,
		Content:   yaml,
		Kind:      "",
		Cluster:   cluster,
		IsGlobal:  false,
		CreatedBy: username,
	}
	// 保存自定义模板对象
	err := ct.Save(nil)
	// 若保存失败，返回错误
	if err != nil {
		return nil, err
	}
	// 若保存成功，返回包含成功消息的调用结果
	return tools.TextResult("保存成功", nil)
}

// GetMcpSSEServer 创建并返回一个集成了“保存K8s YAML模板”工具的MCP SSE服务器实例，支持基于JWT的用户身份提取与Gin框架适配。
func GetMcpSSEServer(basePath string) *server.SSEServer {
	sc := createServerConfig(basePath)
	serv := mcp.GetMCPServerWithOption(sc)
	serv.AddTool(SaveYamlTemplateTool(), SaveYamlTemplateToolHandler)
	return mcp.GetMCPSSEServerWithServerAndOption(serv, sc)
}

// adapt 将标准的 http.Handler 适配为 Gin 框架可用的处理函数。
func adapt(fn func() http.Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		handler := fn()
		handler.ServeHTTP(c.Writer, c.Request)
	}
}

// 手动提取路径中 key（/mcp/k8m/:key/sse）
func extractKey(path string) string {
	parts := strings.Split(path, "/")
	endpoints := []string{"sse", "message"}
	if len(parts) >= 5 && parts[1] == "mcp" && parts[2] == "k8m" && slice.Contain(endpoints, parts[4]) {
		return parts[3]
	}
	return ""
}
