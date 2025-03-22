package service

import (
	"context"
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/mcp"
	"github.com/weibaohui/k8m/pkg/models"
	"k8s.io/klog/v2"
)

type mcpService struct {
	host *mcp.MCPHost
}

func (m *mcpService) Init() {
	if m.host == nil {
		m.host = mcp.NewMCPHost()
	}
	m.run()
}
func (m *mcpService) Host() *mcp.MCPHost {
	return m.host
}
func (m *mcpService) AddServer(server models.MCPServerConfig) {
	// 将server转换为mcp.ServerConfig
	serverConfig := mcp.ServerConfig{
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
		err := m.host.ConnectServer(context.Background(), server.Name)
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
		serverConfig := mcp.ServerConfig{
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
			err := m.host.ConnectServer(context.Background(), server.Name)
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
	serverConfig := mcp.ServerConfig{
		Name:    server.Name,
		URL:     server.URL,
		Enabled: server.Enabled,
	}
	m.host.RemoveServer(serverConfig)
}
func (m *mcpService) run() {

	var mcpServers []models.MCPServerConfig
	err := dao.DB().Model(&models.MCPServerConfig{}).Find(&mcpServers).Error
	if err != nil {
		return
	}
	m.AddServers(mcpServers)

	// 启动定期ping检查
	go func() {
		ticker := time.NewTicker(30 * time.Second) // 每30秒执行一次
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				status := m.host.PingAll(context.Background())
				for serverName, serverStatus := range status {
					if serverStatus.LastPingSuccess {
						klog.V(6).Infof("Server %s is healthy, last ping time: %v", serverName, serverStatus.LastPingTime)
					} else {
						klog.V(6).Infof("Server %s is unhealthy, last ping time: %v, error: %s", serverName, serverStatus.LastPingTime, serverStatus.LastError)
					}
				}
			case <-context.Background().Done():
				return
			}
		}
	}()

}

func (m *mcpService) RemoveServerById(server models.MCPServerConfig) {
	m.host.RemoveServerById(server.ID)
}

func (m *mcpService) UpdateServer(entity models.MCPServerConfig) {
	m.RemoveServerById(entity)
	m.AddServer(entity)
}
