package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sashabaranov/go-openai"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/ai"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/models"
	"k8s.io/klog/v2"
)

// ServerConfig 服务器配置
type ServerConfig struct {
	ID      uint   `json:"id"`
	URL     string `json:"url,omitempty"`
	Name    string `json:"name,omitempty"`
	Enabled bool   `json:"enabled,omitempty"`
}

// MCPHost MCP服务器管理器
type MCPHost struct {
	configs map[string]ServerConfig
	mutex   sync.RWMutex
	// 记录每个服务器的工具列表
	Tools map[string][]mcp.Tool
	// 记录每个服务器的资源能力
	Resources map[string][]mcp.Resource
	// 记录每个服务器的提示能力
	Prompts map[string][]mcp.Prompt

	buffer    []*models.MCPToolLog
	bufferMux sync.Mutex
	ticker    *time.Ticker
	stopChan  chan bool
}
type MCPServer struct {
	ServerConfig
	Config            ServerConfig          `json:"config,omitempty"`
	Tools             []mcp.Tool            `json:"tools,omitempty"`
	Resources         []mcp.Resource        `json:"resources,omitempty"`
	Prompts           []mcp.Prompt          `json:"prompts,omitempty"`
	InitializeResults *mcp.InitializeResult `json:"initialize_results,omitempty"`
}

// NewMCPHost 创建新的MCP管理器
func NewMCPHost() *MCPHost {
	host := &MCPHost{
		configs:   make(map[string]ServerConfig),
		Tools:     make(map[string][]mcp.Tool),
		Resources: make(map[string][]mcp.Resource),
		Prompts:   make(map[string][]mcp.Prompt),

		buffer:   make([]*models.MCPToolLog, 0, 100),
		ticker:   time.NewTicker(2 * time.Second),
		stopChan: make(chan bool),
	}
	return host
}

func (m *MCPHost) ListServers() []MCPServer {

	// 创建结果切片
	var servers []MCPServer

	// 遍历所有配置，转换为MCPServer结构
	for name, config := range m.configs {
		server := MCPServer{
			ServerConfig: config,
			Config:       config,
			Tools:        m.Tools[name],
			Resources:    m.Resources[name],
			Prompts:      m.Prompts[name],
		}
		servers = append(servers, server)
	}
	slice.SortBy(servers, func(a, b MCPServer) bool {
		return a.Config.Name < b.Config.Name
	})
	return servers

}

// AddServer 添加服务器配置
func (m *MCPHost) AddServer(config ServerConfig) error {
	m.RemoveServer(config)
	m.mutex.Lock()
	m.configs[config.Name] = config
	m.mutex.Unlock()
	return nil
}

// SyncServerCapabilities 同步服务器的工具、资源和提示能力
func (m *MCPHost) SyncServerCapabilities(ctx context.Context, serverName string) error {

	// 获取服务器能力
	tools, err := m.GetTools(ctx, serverName)
	if err != nil {
		klog.V(6).Infof("failed to get tools for %s: %v", serverName, err)
	}

	resources, err := m.GetResources(ctx, serverName)
	if err != nil {
		klog.V(6).Infof("failed to get resources for %s: %v", serverName, err)
	}

	prompts, err := m.GetPrompts(ctx, serverName)
	if err != nil {
		klog.V(6).Infof("failed to get prompts for %s: %v", serverName, err)
	}

	// 只在更新共享资源时加锁
	m.mutex.Lock()
	m.Tools[serverName] = tools
	m.Resources[serverName] = resources
	m.Prompts[serverName] = prompts
	m.mutex.Unlock()
	klog.V(6).Infof("同步服务器能力 [%s] 工具:%d 资源:%d 提示:%d", serverName, len(tools), len(resources), len(prompts))
	return nil
}

// ConnectServer 连接到指定服务器
func (m *MCPHost) ConnectServer(ctx context.Context, serverName string) error {
	config, exists := m.configs[serverName]

	if !exists {
		return fmt.Errorf("server config not found: %s", serverName)
	}

	if !config.Enabled {
		return fmt.Errorf("server is disabled: %s", serverName)
	}

	// 在锁外同步服务器能力
	if err := m.SyncServerCapabilities(context.Background(), serverName); err != nil {
		return fmt.Errorf("failed to sync server capabilities for %s: %v", serverName, err)
	}

	return nil
}

// GetClient 获取指定服务器的客户端
func (m *MCPHost) GetClient(ctx context.Context, serverName string) (*client.Client, error) {

	// 获取配置信息
	config, exists := m.configs[serverName]
	if !exists {
		return nil, fmt.Errorf("server config not found: %s", serverName)
	}

	username, _ := m.getUserRoleFromMCPCtx(ctx)
	jwt, err := UserService().GenerateJWTTokenOnlyUserName(username, time.Hour*1)
	if err != nil {
		return nil, fmt.Errorf("failed to generate JWT token for %s: %v", serverName, err)
	}
	// 执行时携带用户名、角色信息
	newCli, err := client.NewSSEMCPClient(config.URL, client.WithHeaders(map[string]string{
		"Authorization": jwt,
	}))
	// klog.V(6).Infof("访问MCP 服务器 [%s:%s] 携带信息%s %s", serverName, config.URL, username, role)
	if err != nil {
		return nil, fmt.Errorf("failed to create new client for %s: %v", serverName, err)
	}

	if err = newCli.Start(ctx); err != nil {
		newCli.Close()
		return nil, fmt.Errorf("failed to start new client for %s: %v", serverName, err)
	}

	//  初始化客户端
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "multi-server-client",
		Version: "1.0.0",
	}

	_, err = newCli.Initialize(ctx, initRequest)
	if err != nil {
		newCli.Close()
		return nil, fmt.Errorf("failed to initialize new client for %s: %v", serverName, err)
	}

	return newCli, nil

}

func (m *MCPHost) getUserRoleFromMCPCtx(ctx context.Context) (string, string) {
	// 重新连接
	username := ""
	if usernameVal, ok := ctx.Value(constants.JwtUserName).(string); ok {
		username = usernameVal
	}
	role := ""
	if roleVal, ok := ctx.Value(constants.JwtUserRole).(string); ok {
		role = roleVal
	}
	return username, role
}

// Close 关闭所有连接
func (m *MCPHost) Close() {
	m.stopChan <- true
}

func (m *MCPHost) GetAllTools(ctx context.Context) []openai.Tool {
	if len(m.Tools) == 0 {
		return nil
	}
	// 从所有可用的MCP服务器收集工具列表
	var allTools []openai.Tool
	// 遍历所有服务器获取工具
	for serverName, tools := range m.Tools {
		for _, tool := range tools {
			allTools = append(allTools, openai.Tool{
				Type: openai.ToolTypeFunction,
				Function: &openai.FunctionDefinition{
					// 在工具名称中添加服务器标识
					Name:        utils.BuildMCPToolName(tool.Name, serverName),
					Description: tool.Name,
					Parameters:  tool.InputSchema,
				},
			})
		}

	}
	return allTools
}

// GetTools 获取指定服务器的工具列表
func (m *MCPHost) GetTools(ctx context.Context, serverName string) ([]mcp.Tool, error) {
	cli, err := m.GetClient(ctx, serverName)
	if err != nil {
		return nil, err
	}

	toolsRequest := mcp.ListToolsRequest{}
	toolsResult, err := cli.ListTools(ctx, toolsRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to get tools from server %s: %v", serverName, err)
	}

	return toolsResult.Tools, nil
}

// GetResources 获取指定服务器的资源能力
func (m *MCPHost) GetResources(ctx context.Context, serverName string) ([]mcp.Resource, error) {
	cli, err := m.GetClient(ctx, serverName)
	if err != nil {
		return nil, err
	}
	req := mcp.ListResourcesRequest{}
	result, err := cli.ListResources(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get resources from server %s: %v", serverName, err)
	}

	return result.Resources, nil
}

// GetPrompts 获取指定服务器的提示能力
func (m *MCPHost) GetPrompts(ctx context.Context, serverName string) ([]mcp.Prompt, error) {
	cli, err := m.GetClient(ctx, serverName)
	if err != nil {
		return nil, err
	}
	req := mcp.ListPromptsRequest{}
	result, err := cli.ListPrompts(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get prompts from server %s: %v", serverName, err)
	}

	return result.Prompts, nil
}

func (m *MCPHost) RemoveServer(config ServerConfig) {
	m.mutex.Lock()

	// 删除服务器配置
	delete(m.configs, config.Name)
	// 删除服务器的工具、资源和提示能力
	delete(m.Tools, config.Name)
	delete(m.Resources, config.Name)
	delete(m.Prompts, config.Name)
	m.mutex.Unlock()
}

func (m *MCPHost) RemoveServerById(id uint) {
	for _, cfg := range m.configs {
		if cfg.ID == id {
			m.RemoveServer(cfg)
		}
	}
}

// GetServerNameByToolName 根据工具名称获取对应的服务器名称
// 如果多个服务器都提供了相同的工具，返回第一个找到的服务器名称，有一定的随机性
// 如果没有找到对应的服务器，返回空字符串
func (m *MCPHost) GetServerNameByToolName(toolName string) string {

	for serverName, tools := range m.Tools {
		for _, tool := range tools {
			if tool.Name == toolName {
				return serverName
			}
		}
	}
	return ""
}

// LogToolExecution 记录工具执行日志
func (m *MCPHost) LogToolExecution(ctx context.Context, toolName, serverName string, parameters interface{}, result models.MCPToolCallResult, executeTime int64) {

	log := &models.MCPToolLog{
		ToolName:    toolName,
		ServerName:  serverName,
		Parameters:  utils.ToJSON(parameters),
		Result:      result.Result,
		ExecuteTime: executeTime,
		CreatedAt:   time.Now(),
		Error:       result.Error,
	}

	username, _ := m.getUserRoleFromMCPCtx(ctx)
	log.CreatedBy = username

	prompt := ""
	if promptVal, ok := ctx.Value("prompt").(string); ok {
		prompt = promptVal
	}
	log.Prompt = prompt
	if log.Result == "" && log.Error != "" {
		log.Result = log.Error
	}

	dao.DB().Create(log)
}

func (m *MCPHost) ProcessWithOpenAI(ctx context.Context, ai ai.IAI, prompt string) (string, []models.MCPToolCallResult, error) {

	// 创建带有工具的聊天完成请求
	tools := m.GetAllTools(ctx)
	ai.SetTools(tools)
	toolCalls, content, err := ai.GetCompletionWithTools(ctx, prompt)
	if err != nil {
		return "", nil, err
	}

	results := m.ExecTools(ctx, toolCalls)

	return content, results, nil

}

func (m *MCPHost) ExecTools(ctx context.Context, toolCalls []openai.ToolCall) []models.MCPToolCallResult {
	// 存储所有工具调用的结果
	var results []models.MCPToolCallResult

	// 处理工具调用
	if toolCalls != nil {
		for _, toolCall := range toolCalls {
			startTime := time.Now()

			fullToolName := toolCall.Function.Name
			klog.V(6).Infof("Tool Name: %s\n", fullToolName)
			arguments := toolCall.Function.Arguments
			arguments = clean(arguments)
			klog.V(6).Infof("Tool Arguments: %s\n", arguments)

			result := models.MCPToolCallResult{
				ToolName: fullToolName,
			}

			// 解析参数
			var args map[string]interface{}
			if arguments != "" && arguments != "{}" && arguments != "null" {

				if err := json.Unmarshal([]byte(arguments), &args); err != nil {
					result.Error = fmt.Sprintf("failed to parse tool arguments: %v", err)
					klog.V(6).Infof("参数解析Error: %s\n", result.Error)
					results = append(results, result)
					continue
				}
			}

			result.Parameters = args

			var cli *client.Client
			var toolName, serverName string
			var err error
			if strings.Contains(fullToolName, "@") {
				// 如果识别的ToolName包含@，则解析ToolName
				toolName, serverName, _ = utils.ParseMCPToolName(fullToolName)
			} else {
				toolName = fullToolName
				serverName = m.GetServerNameByToolName(toolName)
			}
			klog.V(6).Infof("解析ToolName: %s, ServerName: %s\n", toolName, serverName)
			if serverName == "" {
				// 解析失败，尝试直接用toolName
				result.Error = fmt.Sprintf("根据Tool名称 %s 解析MCP Server 名称失败: %v", fullToolName, err)
				results = append(results, result)
				continue
			}
			klog.V(6).Infof("解析ToolName: %s, ServerName: %s\n", toolName, serverName)
			// 执行工具调用
			callRequest := mcp.CallToolRequest{}
			callRequest.Params.Name = toolName
			callRequest.Params.Arguments = args
			klog.V(6).Infof("执行工具调用: %s\n", utils.ToJSON(callRequest))
			cli, err = m.GetClient(ctx, serverName)

			if err != nil {
				klog.V(6).Infof("获取MCP Client 失败: %v\n", err)
				result.Error = fmt.Sprintf("获取MCP Client 失败: %v", err)
				results = append(results, result)
				continue
			}
			// 执行工具
			callResult, err := cli.CallTool(ctx, callRequest)
			_ = cli.Close()
			// 记录执行日志
			executeTime := time.Since(startTime).Milliseconds()
			if err != nil {
				klog.V(6).Infof("工具执行失败: %v\n", err)
				result.Error = fmt.Sprintf("工具执行失败: %v", err)
				results = append(results, result)
				m.LogToolExecution(ctx, toolName, serverName, args, result, executeTime)
				continue
			}

			// 处理工具执行结果
			if len(callResult.Content) > 0 {
				if textContent, ok := callResult.Content[0].(mcp.TextContent); ok {
					result.Result = textContent.Text
				}
			}
			results = append(results, result)

			m.LogToolExecution(ctx, toolName, serverName, args, result, executeTime)
		}
	}
	return results
}

// clean 移除参数字符串中的多余空白字符、换行和制表符，并将空的 JSON 对象格式（"{}" 或 "{}}"）标准化为空字符串。
func clean(arguments string) string {
	arguments = strings.TrimSpace(arguments)
	arguments = strings.ReplaceAll(arguments, "\n", "")
	arguments = strings.ReplaceAll(arguments, "\t", "")
	if arguments == "{}}" {
		arguments = ""
	}
	if arguments == "{}" {
		arguments = ""
	}
	return arguments
}
