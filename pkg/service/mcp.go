package service

import (
	"context"
	"encoding/json"
	"time"

	mcp2 "github.com/mark3labs/mcp-go/mcp"
	"github.com/sashabaranov/go-openai"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/models"
	"k8s.io/klog/v2"
)

type mcpService struct {
	host *MCPHost
}

func (m *mcpService) Init() {
	if m.host == nil {
		m.host = NewMCPHost()
	}
	m.Start()
}
func (m *mcpService) Host() *MCPHost {
	return m.host
}
func (m *mcpService) AddServer(server models.MCPServerConfig) {
	// 将server转换为mcp.ServerConfig
	serverConfig := ServerConfig{
		ID:      server.ID,
		Name:    server.Name,
		URL:     server.URL,
		Enabled: server.Enabled,
	}
	err := m.host.AddServer(serverConfig)
	if err != nil {
		klog.V(6).Infof("Failed to add server %s: %v", server.Name, err)
		return
	}

	if server.Enabled {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		err := m.host.ConnectServer(ctx, server.Name)
		if err != nil {
			klog.V(6).Infof("Failed to connect to server %s: %v", server.Name, err)
			return
		}
		klog.V(6).Infof("Successfully connected to server: %s", server.Name)
	}

}
func (m *mcpService) AddServers(servers []models.MCPServerConfig) {
	for _, server := range servers {
		// 将server转换为mcp.ServerConfig
		serverConfig := ServerConfig{
			ID:      server.ID,
			Name:    server.Name,
			URL:     server.URL,
			Enabled: server.Enabled,
		}
		err := m.host.AddServer(serverConfig)
		if err != nil {
			klog.V(6).Infof("Failed to add server %s: %v", server.Name, err)
			continue
		}

		if server.Enabled {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			err := m.host.ConnectServer(ctx, server.Name)
			cancel()
			if err != nil {
				klog.V(6).Infof("Failed to connect to server %s: %v", server.Name, err)
				continue
			}
			klog.V(6).Infof("Successfully connected to server: %s", server.Name)
		}
	}

}
func (m *mcpService) RemoveServer(server models.MCPServerConfig) {
	// 将server转换为mcp.ServerConfig
	serverConfig := ServerConfig{
		Name:    server.Name,
		URL:     server.URL,
		Enabled: server.Enabled,
	}
	m.host.RemoveServer(serverConfig)
}
func (m *mcpService) Start() {

	var mcpServers []models.MCPServerConfig
	err := dao.DB().Model(&models.MCPServerConfig{}).Find(&mcpServers).Error
	if err != nil {
		return
	}
	m.AddServers(mcpServers)
}

func (m *mcpService) RemoveServerById(server models.MCPServerConfig) {
	m.host.RemoveServerById(server.ID)
}

func (m *mcpService) UpdateServer(entity models.MCPServerConfig) {
	m.RemoveServerById(entity)
	m.AddServer(entity)
}

func (m *mcpService) GetTools(entity models.MCPServerConfig) ([]mcp2.Tool, error) {
	ctx := context.Background()
	return m.Host().GetTools(ctx, entity.Name)
}

func (m *mcpService) GetAllEnabledTools() []openai.Tool {

	var tools []models.MCPTool
	err := dao.DB().Model(&models.MCPTool{}).Where("enabled = ?", true).Find(&tools).Error
	if err != nil {
		return nil
	}
	// 从所有可用的MCP服务器收集工具列表
	var allTools []openai.Tool
	// 遍历所有服务器获取工具
	for _, tool := range tools {

		var tis mcp2.ToolInputSchema
		err := json.Unmarshal([]byte(tool.InputSchema), &tis)
		if err != nil {
			continue
		}
		allTools = append(allTools, openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				// 在工具名称中添加服务器标识
				Name:        utils.BuildMCPToolName(tool.Name, tool.ServerName),
				Description: tool.Name,
				// 将工具的输入模式转换为紧凑的JSON格式
				Parameters: tis,
			},
		})
	}
	return allTools

}
