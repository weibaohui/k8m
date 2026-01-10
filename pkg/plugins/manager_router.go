package plugins

import (
	"slices"

	"github.com/go-chi/chi/v5"
	"k8s.io/klog/v2"
)

// RegisterClusterRoutes 某个插件的集群操作相关的路由注册
// 路径/k8s/cluster/<clusterID>/plugins/<pluginName>/xxx
func (m *Manager) RegisterClusterRoutes(api chi.Router) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// 记录路由分组（去重）
	already := slices.Contains(m.apiGroups, api)
	if !already {
		m.apiGroups = append(m.apiGroups, api)
	}
	// 为已启用插件注册路由
	for name, mod := range m.modules {
		if m.status[name] == StatusRunning && mod.ClusterRouter != nil {
			klog.V(6).Infof("注册插件 集群 路由: %s", name)
			mod.ClusterRouter(api)
		}
	}
}

// RegisterManagementRoutes 某个插件的管理相关的操作的路由注册
// 路径/mgm/plugins/<pluginName>/yyy
func (m *Manager) RegisterManagementRoutes(api chi.Router) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// 记录路由分组（去重）
	already := slices.Contains(m.apiGroups, api)
	if !already {
		m.apiGroups = append(m.apiGroups, api)
	}
	// 为已启用插件注册路由
	for name, mod := range m.modules {
		if m.status[name] == StatusRunning && mod.ManagementRouter != nil {
			klog.V(6).Infof("注册插件 管理 路由: %s", name)
			mod.ManagementRouter(api)
		}
	}
}

// RegisterPluginAdminRoutes 某个插件的管理相关的操作的路由注册
// 路径/admin/plugins/<pluginName>/yyy
func (m *Manager) RegisterPluginAdminRoutes(api chi.Router) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// 记录路由分组（去重）
	already := slices.Contains(m.apiGroups, api)
	if !already {
		m.apiGroups = append(m.apiGroups, api)
	}
	// 为已启用插件注册路由
	for name, mod := range m.modules {
		if m.status[name] == StatusRunning && mod.PluginAdminRouter != nil {
			klog.V(6).Infof("注册插件 插件管理 路由: %s", name)
			mod.PluginAdminRouter(api)
		}
	}
}

// RegisterRootRoutes 某个插件的根路由注册
// 路径 /
func (m *Manager) RegisterRootRoutes(root chi.Router) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// 记录路由分组（去重）
	already := slices.Contains(m.apiGroups, root)
	if !already {
		m.apiGroups = append(m.apiGroups, root)
	}
	// 为已启用插件注册路由
	for name, mod := range m.modules {
		if m.status[name] == StatusRunning && mod.RootRouter != nil {
			klog.V(6).Infof("注册插件 根 路由: %s", name)
			mod.RootRouter(root)
		}
	}
}
