package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	mcp2 "github.com/mark3labs/mcp-go/mcp"
	"github.com/sashabaranov/go-openai"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	uModels "github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/plugins/modules/mcp/models"

	"gorm.io/gorm"

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
func (m *mcpService) AddServer(ctx context.Context, server models.MCPServerConfig) {
	// 将server转换为mcp.ServerConfig
	serverConfig := ServerConfig{
		ID:      server.ID,
		Name:    server.Name,
		URL:     server.URL,
		Enabled: server.Enabled,
	}
	err := m.host.AddServer(serverConfig)
	if err != nil {
		klog.V(6).Infof("添加服务器 %s 失败: %v", server.Name, err)
		return
	}

	if server.Enabled {
		ctxc, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		err := m.host.ConnectServer(ctxc, server.Name)
		if err != nil {
			klog.V(6).Infof("连接服务器 %s 失败: %v", server.Name, err)
			return
		}
		klog.V(6).Infof("成功连接到服务器: %s", server.Name)
	}

}
func (m *mcpService) AddServers(ctx context.Context, servers []models.MCPServerConfig) {
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
			klog.V(6).Infof("添加服务器 %s 失败: %v", server.Name, err)
			continue
		}

		if server.Enabled {

			ctxc, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()
			err := m.host.ConnectServer(ctxc, server.Name)

			if err != nil {
				klog.V(6).Infof("连接服务器 %s 失败: %v", server.Name, err)
				continue
			}
			klog.V(6).Infof("成功连接到服务器: %s", server.Name)
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
	ctx := amis.GetContextForAdmin()
	m.AddServers(ctx, mcpServers)
	klog.V(6).Infof("成功启动 MCP 服务，共 %d 个服务器", len(mcpServers))
}

func (m *mcpService) RemoveServerById(server models.MCPServerConfig) {
	m.host.RemoveServerById(server.ID)
}

func (m *mcpService) UpdateServer(ctx context.Context, entity models.MCPServerConfig) {
	m.RemoveServerById(entity)
	m.AddServer(ctx, entity)
}

func (m *mcpService) GetTools(ctx context.Context, entity models.MCPServerConfig) ([]mcp2.Tool, error) {
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
				Name:        BuildMCPToolName(tool.Name, tool.ServerName),
				Description: tool.Name,
				// 将工具的输入模式转换为紧凑的JSON格式
				Parameters: tis,
			},
		})
	}
	return allTools

}

func (m *mcpService) GetUserByMCPKey(mcpKey string) (string, error) {
	params := &dao.Params{}
	md := &models.McpKey{}
	queryFunc := func(db *gorm.DB) *gorm.DB {
		return db.Select("username").Where(" mcp_key = ?", mcpKey)
	}
	item, err := md.GetOne(params, queryFunc)
	if err != nil {
		return "", err
	}

	if item.Username == "" {
		return "", errors.New("username is empty")
	}

	// 检测用户是否被禁用
	user := &uModels.User{}
	disabled, err := user.IsDisabled(item.Username)
	if err != nil {
		return "", err
	}
	if disabled {
		return "", fmt.Errorf("用户[%s]被禁用", item.Username)
	}
	return item.Username, nil
}

// BuildMCPToolName 构建完整的工具名称
func BuildMCPToolName(toolName, serverName string) string {
	return fmt.Sprintf("%s@%s", toolName, serverName)
}

// ParseMCPToolName 从完整的工具名称中解析出服务器名称
func ParseMCPToolName(fullToolName string) (toolName, serverName string, err error) {
	lastIndex := strings.LastIndex(fullToolName, "@")
	if lastIndex == -1 {
		return "", "", fmt.Errorf("invalid tool name format: %s", fullToolName)
	}
	return fullToolName[:lastIndex], fullToolName[lastIndex+1:], nil
}
