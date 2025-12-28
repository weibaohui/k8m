package plugins

import (
	"slices"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"
)

// RegisterClusterRoutes 某个插件的集群操作相关的路由注册
func (m *Manager) RegisterClusterRoutes(api *gin.RouterGroup) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// 记录路由分组（去重）
	already := slices.Contains(m.apiGroups, api)
	if !already {
		m.apiGroups = append(m.apiGroups, api)
	}
	// 为已启用插件注册路由
	for name, mod := range m.modules {
		if m.status[name] == StatusEnabled && mod.ClusterRouter != nil {
			klog.V(6).Infof("注册插件路由: %s", name)
			mod.ClusterRouter(api)
		}
	}
}

// RegisterManagementRoutes 某个插件的管理相关的操作的路由注册
func (m *Manager) RegisterManagementRoutes(api *gin.RouterGroup) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// 记录路由分组（去重）
	already := slices.Contains(m.apiGroups, api)
	if !already {
		m.apiGroups = append(m.apiGroups, api)
	}
	// 为已启用插件注册路由
	for name, mod := range m.modules {
		if m.status[name] == StatusEnabled && mod.ManagementRouter != nil {
			klog.V(6).Infof("注册插件路由: %s", name)
			mod.ManagementRouter(api)
		}
	}
}

// RegisterPluginAdminRoutes 某个插件的管理相关的操作的路由注册
func (m *Manager) RegisterPluginAdminRoutes(api *gin.RouterGroup) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// 记录路由分组（去重）
	already := slices.Contains(m.apiGroups, api)
	if !already {
		m.apiGroups = append(m.apiGroups, api)
	}
	// 为已启用插件注册路由
	for name, mod := range m.modules {
		if m.status[name] == StatusEnabled && mod.PluginAdminRouter != nil {
			klog.V(6).Infof("注册插件路由: %s", name)
			mod.PluginAdminRouter(api)
		}
	}
}
