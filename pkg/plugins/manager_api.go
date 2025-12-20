package plugins

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"k8s.io/klog/v2"
)

// RegisterAdminRoutes 注册插件的管理员路由
// 管理员路由通常用于插件的配置、管理和操作接口，需要较高的权限才能访问。
// 提供功能：
// 1. 插件列表（显示Meta信息与状态）
// 2. 安装插件
// 3. 卸载插件
// 4. 页面Schema（AMIS）
func (m *Manager) RegisterAdminRoutes(admin *gin.RouterGroup) {
	grp := admin.Group("/plugin")

	// 列出所有已注册插件的Meta和状态
	grp.GET("/list", func(c *gin.Context) {
		m.mu.RLock()
		defer m.mu.RUnlock()

		type PluginItem struct {
			Name        string `json:"name"`
			Title       string `json:"title"`
			Version     string `json:"version"`
			Description string `json:"description"`
			Status      string `json:"status"`
		}

		items := make([]PluginItem, 0, len(m.modules))
		for name, mod := range m.modules {
			status := m.status[name]
			items = append(items, PluginItem{
				Name:        mod.Meta.Name,
				Title:       mod.Meta.Title,
				Version:     mod.Meta.Version,
				Description: mod.Meta.Description,
				Status:      statusToCN(status),
			})
		}
		klog.V(6).Infof("获取插件列表，共计%d个", len(items))
		amis.WriteJsonListWithTotal(c, int64(len(items)), items)
	})

	// 安装插件
	grp.POST("/install/:name", func(c *gin.Context) {
		name := c.Param("name")
		klog.V(6).Infof("安装插件请求: %s", name)
		if err := m.Install(name); err != nil {
			amis.WriteJsonError(c, err)
			return
		}
		amis.WriteJsonOK(c)
	})

	// 卸载插件
	grp.POST("/uninstall/:name", func(c *gin.Context) {
		name := c.Param("name")
		klog.V(6).Infof("卸载插件请求: %s", name)
		if err := m.Uninstall(name); err != nil {
			amis.WriteJsonError(c, err)
			return
		}
		amis.WriteJsonOK(c)
	})
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

