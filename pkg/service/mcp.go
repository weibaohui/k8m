package service

import (
	"context"
	"time"

	"github.com/weibaohui/k8m/pkg/mcp"
	"k8s.io/klog/v2"
)

type mcpService struct {
	host *mcp.MCPHost
}

func (m *mcpService) Init() {
	if m.host == nil {
		m.host = mcp.NewMCPHost()
	}
	m.Run()
}
func (m *mcpService) Host() *mcp.MCPHost {
	return m.host
}
func (m *mcpService) Run() {

	// 创建MCP管理器

	// 添加服务器配置
	servers := []mcp.ServerConfig{
		{
			Name:    "server1",
			URL:     "http://localhost:9292/sse",
			Enabled: true,
		},
		{
			Name:    "server2",
			URL:     "http://localhost:9293/sse",
			Enabled: true,
		},
		{
			Name:    "github",
			URL:     "https://mcp.composio.dev/github/repulsive-quaint-plumber-9mTBRR",
			Enabled: true,
		}, {
			Name:    "dynamics365",
			URL:     "https://mcp.composio.dev/dynamics365/repulsive-quaint-plumber-9mTBRR",
			Enabled: true,
		}, {
			Name:    "time",
			URL:     "https://router.mcp.so/sse/po6bz3m8iuv5qg",
			Enabled: true,
		}, {
			Name:    "Fetch",
			URL:     "https://router.mcp.so/sse/ji513cm8iv20ga",
			Enabled: true,
		},
	}

	// 添加并连接服务器
	ctx := context.Background()
	for _, server := range servers {
		if err := m.host.AddServer(server); err != nil {
			klog.V(6).Infof("Failed to add server %s: %v", server.Name, err)
			continue
		}

		if err := m.host.ConnectServer(ctx, server.Name); err != nil {
			klog.V(6).Infof("Failed to connect to server %s: %v", server.Name, err)
			continue
		}

		klog.V(6).Infof("Successfully connected to server: %s", server.Name)
	}

	// 启动定期ping检查
	go func() {
		ticker := time.NewTicker(30 * time.Second) // 每30秒执行一次
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				status := m.host.PingAll(ctx)
				for serverName, serverStatus := range status {
					if serverStatus.LastPingSuccess {
						klog.V(6).Infof("Server %s is healthy, last ping time: %v", serverName, serverStatus.LastPingTime)
					} else {
						klog.V(6).Infof("Server %s is unhealthy, last ping time: %v, error: %s", serverName, serverStatus.LastPingTime, serverStatus.LastError)
					}
				}
			case <-ctx.Done():
				return
			}
		}
	}()

}
