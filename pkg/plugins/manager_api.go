package plugins

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"k8s.io/klog/v2"
)

// PluginItemVO 插件列表展示结构体
// 用于在管理员接口中返回插件的基础信息与当前状态
type PluginItemVO struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

// RegisterAdminRoutes 注册插件的管理员路由
// 管理员路由通常用于插件的配置、管理和操作接口，需要较高的权限才能访问。
// 提供功能：
// 1. 插件列表（显示Meta信息与状态）
// 2. 安装插件
// 3. 卸载插件
func (m *Manager) RegisterAdminRoutes(admin *gin.RouterGroup) {
	grp := admin.Group("/plugin")

	// 列出所有已注册插件的Meta和状态
	grp.GET("/list", m.ListPlugins)
	// 获取已启用插件的菜单数据
	grp.GET("/menus", m.ListPluginMenus)
	// 安装插件
	grp.POST("/install/:name", m.InstallPlugin)
	// 启用插件
	grp.POST("/enable/:name", m.EnablePlugin)
	// 禁用插件
	grp.POST("/disable/:name", m.DisablePlugin)
	// 卸载插件
	grp.POST("/uninstall/:name", m.UninstallPlugin)
}

// ListPlugins 获取所有已注册插件的Meta与状态
// 返回插件名称、标题、版本、描述及当前状态（中文）
func (m *Manager) ListPlugins(c *gin.Context) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	items := make([]PluginItemVO, 0, len(m.modules))
	for name, mod := range m.modules {
		status := m.status[name]
		items = append(items, PluginItemVO{
			Name:        mod.Meta.Name,
			Title:       mod.Meta.Title,
			Version:     mod.Meta.Version,
			Description: mod.Meta.Description,
			Status:      statusToCN(status),
		})
	}
	klog.V(6).Infof("获取插件列表，共计%d个", len(items))
	amis.WriteJsonListWithTotal(c, int64(len(items)), items)
}

// ListPluginMenus 获取所有已启用插件的菜单定义
// 返回前端可直接使用的菜单JSON（与前端 MenuItem 结构一致）
func (m *Manager) ListPluginMenus(c *gin.Context) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	type MenuVO struct {
		Key         string   `json:"key,omitempty"`
		Title       string   `json:"title"`
		Icon        string   `json:"icon,omitempty"`
		URL         string   `json:"url,omitempty"`
		EventType   string   `json:"eventType,omitempty"`
		CustomEvent string   `json:"customEvent,omitempty"`
		Order       float64  `json:"order,omitempty"`
		Children    []MenuVO `json:"children,omitempty"`
		Show        string   `json:"show,omitempty"`
		Permission  string   `json:"permission,omitempty"`
	}

	var result []MenuVO
	for name, mod := range m.modules {
		if m.status[name] != StatusEnabled {
			continue
		}
		for _, menu := range mod.Menus {
			result = append(result, MenuVO{
				Key:         menu.Key,
				Title:       menu.Title,
				Icon:        menu.Icon,
				URL:         menu.URL,
				EventType:   menu.EventType,
				CustomEvent: menu.CustomEvent,
				Order:       menu.Order,
				Show:        menu.Show,
				Permission:  menu.Permission,
			})
		}
	}
	klog.V(6).Infof("获取插件菜单列表，共计%d个", len(result))
	amis.WriteJsonData(c, result)
}

// InstallPlugin 安装指定名称的插件
// 路径参数为插件名，安装失败时返回错误
func (m *Manager) InstallPlugin(c *gin.Context) {
	name := c.Param("name")
	klog.V(6).Infof("安装插件请求: %s", name)
	if err := m.Install(name); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

// EnablePlugin 启用指定名称的插件
// 路径参数为插件名，启用失败时返回错误
func (m *Manager) EnablePlugin(c *gin.Context) {
	name := c.Param("name")
	klog.V(6).Infof("启用插件请求: %s", name)
	if err := m.Enable(name); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

// UninstallPlugin 卸载指定名称的插件
// 路径参数为插件名，卸载失败时返回错误
func (m *Manager) UninstallPlugin(c *gin.Context) {
	name := c.Param("name")
	klog.V(6).Infof("卸载插件请求: %s", name)
	if err := m.Uninstall(name); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

// DisablePlugin 禁用指定名称的插件
// 路径参数为插件名，禁用失败时返回错误
func (m *Manager) DisablePlugin(c *gin.Context) {
	name := c.Param("name")
	klog.V(6).Infof("禁用插件请求: %s", name)
	if err := m.Disable(name); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

// statusToCN 状态转中文字符串
func statusToCN(s Status) string {
	switch s {
	case StatusDiscovered:
		return "已发现"
	case StatusInstalled:
		return "已安装"
	case StatusEnabled:
		return "已启用"
	case StatusDisabled:
		return "已禁用"
	default:
		return "未知"
	}
}
