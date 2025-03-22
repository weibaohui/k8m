package mcp

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sashabaranov/go-openai"
	"k8s.io/klog/v2"
)

// ServerConfig 服务器配置
type ServerConfig struct {
	ID      uint   `json:"id"`
	URL     string `json:"url,omitempty"`
	Name    string `json:"name,omitempty"`
	Enabled bool   `json:"enabled,omitempty"`
}

// ServerStatus 服务器状态记录
type ServerStatus struct {
	LastPingTime    *time.Time `json:"last_ping_time,omitempty"`
	LastPingSuccess bool       `json:"last_ping_success,omitempty"`
	LastError       string     `json:"last_error,omitempty"`
}

// MCPHost MCP服务器管理器
type MCPHost struct {
	clients map[string]*client.SSEMCPClient
	configs map[string]ServerConfig
	mutex   sync.RWMutex
	// 记录每个服务器的工具列表
	Tools map[string][]mcp.Tool
	// 记录每个服务器的资源能力
	Resources map[string][]mcp.Resource
	// 记录每个服务器的提示能力
	Prompts           map[string][]mcp.Prompt
	InitializeResults map[string]*mcp.InitializeResult
	// 记录服务器状态
	serverStatus map[string]ServerStatus
}
type MCPServer struct {
	ServerConfig
	Config            ServerConfig          `json:"config,omitempty"`
	Tools             []mcp.Tool            `json:"tools,omitempty"`
	Resources         []mcp.Resource        `json:"resources,omitempty"`
	Prompts           []mcp.Prompt          `json:"prompts,omitempty"`
	InitializeResults *mcp.InitializeResult `json:"initialize_results,omitempty"`
	ServerStatus      ServerStatus          `json:"server_status,omitempty"`
}

// NewMCPHost 创建新的MCP管理器
func NewMCPHost() *MCPHost {
	return &MCPHost{
		clients:           make(map[string]*client.SSEMCPClient),
		configs:           make(map[string]ServerConfig),
		Tools:             make(map[string][]mcp.Tool),
		Resources:         make(map[string][]mcp.Resource),
		Prompts:           make(map[string][]mcp.Prompt),
		InitializeResults: make(map[string]*mcp.InitializeResult),
		serverStatus:      make(map[string]ServerStatus),
	}
}

func (m *MCPHost) ListServers() []MCPServer {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 创建结果切片
	var servers []MCPServer

	// 遍历所有配置，转换为MCPServer结构
	for name, config := range m.configs {
		server := MCPServer{
			ServerConfig:      config,
			Config:            config,
			Tools:             m.Tools[name],
			Resources:         m.Resources[name],
			Prompts:           m.Prompts[name],
			InitializeResults: m.InitializeResults[name],
			ServerStatus:      m.serverStatus[name],
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

	return nil
}

// ConnectServer 连接到指定服务器
func (m *MCPHost) ConnectServer(ctx context.Context, serverName string) error {
	// 获取配置信息时加锁
	m.mutex.Lock()
	config, exists := m.configs[serverName]
	m.mutex.Unlock()

	if !exists {
		return fmt.Errorf("server config not found: %s", serverName)
	}

	if !config.Enabled {
		return fmt.Errorf("server is disabled: %s", serverName)
	}

	cli, err := client.NewSSEMCPClient(config.URL)
	if err != nil {
		return fmt.Errorf("failed to create client for %s: %v", serverName, err)
	}

	if err := cli.Start(ctx); err != nil {
		cli.Close()
		return fmt.Errorf("failed to start client for %s: %v", serverName, err)
	}

	// 初始化客户端
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "multi-server-client",
		Version: "1.0.0",
	}
	result, err := cli.Initialize(ctx, initRequest)
	if err != nil {
		cli.Close()
		return fmt.Errorf("failed to initialize client for %s: %v", serverName, err)
	}

	// 更新共享资源时加锁
	m.mutex.Lock()
	m.clients[serverName] = cli
	m.InitializeResults[serverName] = result
	m.mutex.Unlock()

	// 在锁外同步服务器能力
	if err = m.SyncServerCapabilities(ctx, serverName); err != nil {
		// 如果同步失败，需要清理资源
		cli.Close()
		m.mutex.Lock()
		delete(m.clients, serverName)
		m.mutex.Unlock()
		return fmt.Errorf("failed to sync server capabilities for %s: %v", serverName, err)
	}

	return nil
}

// DisconnectServer 断开与指定服务器的连接
func (m *MCPHost) DisconnectServer(serverName string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if cli, exists := m.clients[serverName]; exists {
		cli.Close()
		delete(m.clients, serverName)
	}
	return nil
}

// GetClient 获取指定服务器的客户端
func (m *MCPHost) GetClient(serverName string) (*client.SSEMCPClient, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	cli, exists := m.clients[serverName]
	if !exists {
		return nil, fmt.Errorf("client not found: %s", serverName)
	}
	return cli, nil
}

// Close 关闭所有连接
func (m *MCPHost) Close() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, cli := range m.clients {
		cli.Close()
	}
	m.clients = make(map[string]*client.SSEMCPClient)
}

// Ping 检测指定服务器的连接状态
func (m *MCPHost) Ping(ctx context.Context, serverName string) error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	cli, exists := m.clients[serverName]
	if !exists {
		return fmt.Errorf("client not found: %s", serverName)
	}

	err := cli.Ping(ctx)
	if err != nil {
		return fmt.Errorf("ping failed for server %s: %v", serverName, err)
	}

	return nil
}

func (m *MCPHost) GetAllTools(ctx context.Context) []openai.Tool {
	// 从所有可用的MCP服务器收集工具列表
	var allTools []openai.Tool
	// 遍历所有服务器获取工具
	m.mutex.RLock()
	for serverName, tools := range m.Tools {
		for _, tool := range tools {
			allTools = append(allTools, openai.Tool{
				Type: openai.ToolTypeFunction,
				Function: &openai.FunctionDefinition{
					// 在工具名称中添加服务器标识
					Name:        buildToolName(tool.Name, serverName),
					Description: tool.Name,
					Parameters:  tool.InputSchema,
				},
			})
		}

	}
	m.mutex.RUnlock()
	return allTools
}

// GetTools 获取指定服务器的工具列表
func (m *MCPHost) GetTools(ctx context.Context, serverName string) ([]mcp.Tool, error) {
	// 获取客户端时加读锁
	m.mutex.RLock()
	cli, exists := m.clients[serverName]
	m.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("client not found: %s", serverName)
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
	// 获取客户端时加读锁
	m.mutex.RLock()
	cli, exists := m.clients[serverName]
	m.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("client not found: %s", serverName)
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
	// 获取客户端时加读锁
	m.mutex.RLock()
	cli, exists := m.clients[serverName]
	m.mutex.RUnlock()

	if !exists {
		return nil, fmt.Errorf("client not found: %s", serverName)
	}

	req := mcp.ListPromptsRequest{}
	result, err := cli.ListPrompts(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get prompts from server %s: %v", serverName, err)
	}

	return result.Prompts, nil
}

// PingAll 检测所有服务器的连接状态
func (m *MCPHost) PingAll(ctx context.Context) map[string]ServerStatus {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// 如果serverStatus为nil，初始化它
	if m.serverStatus == nil {
		m.serverStatus = make(map[string]ServerStatus)
	}

	// 遍历所有客户端进行ping操作
	for serverName, cli := range m.clients {
		now := time.Now()
		status := ServerStatus{
			LastPingTime: &now,
		}

		err := cli.Ping(ctx)
		if err != nil {
			status.LastPingSuccess = false
			status.LastError = err.Error()
			log.Printf("Ping failed for server %s: %v", serverName, err)
		} else {
			status.LastPingSuccess = true
			status.LastError = ""
		}

		m.serverStatus[serverName] = status
	}

	return m.serverStatus
}

// GetServerStatus 获取指定服务器的状态
func (m *MCPHost) GetServerStatus(serverName string) (ServerStatus, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if status, exists := m.serverStatus[serverName]; exists {
		return status, nil
	}
	return ServerStatus{}, fmt.Errorf("server status not found: %s", serverName)
}

// GetAllServerStatus 获取所有服务器的状态
func (m *MCPHost) GetAllServerStatus() map[string]ServerStatus {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	// 创建一个新的map来存储状态的副本
	statusCopy := make(map[string]ServerStatus)
	for k, v := range m.serverStatus {
		statusCopy[k] = v
	}

	return statusCopy
}

func (m *MCPHost) RemoveServer(config ServerConfig) {
	// 断开与服务器的连接
	m.DisconnectServer(config.Name)
	// 删除服务器配置
	delete(m.configs, config.Name)
	// 删除服务器的工具、资源和提示能力
	delete(m.Tools, config.Name)
	delete(m.Resources, config.Name)
	delete(m.Prompts, config.Name)
	delete(m.InitializeResults, config.Name)
	// 删除服务器状态记录
	delete(m.serverStatus, config.Name)

}

func (m *MCPHost) RemoveServerById(id uint) {
	for _, cfg := range m.configs {
		if cfg.ID == id {
			m.RemoveServer(cfg)
		}
	}
}
