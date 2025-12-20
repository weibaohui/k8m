package plugins

import (
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"k8s.io/klog/v2"
)

// Manager 管理插件的注册、安装、启用和禁用
type Manager struct {
	mu      sync.RWMutex
	modules map[string]Module
	status  map[string]Status
}

// Status 插件状态
type Status int

const (
	// StatusDiscovered 已发现
	StatusDiscovered Status = iota
	// StatusInstalled 已安装未启用
	StatusInstalled
	// StatusEnabled 已启用
	StatusEnabled
	// StatusDisabled 已禁用
	StatusDisabled
)

// NewManager 创建并返回插件管理器
func NewManager() *Manager {
	return &Manager{
		modules: make(map[string]Module),
		status:  make(map[string]Status),
	}
}

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

	// 返回AMIS页面Schema，用于管理插件
	grp.GET("/page", func(c *gin.Context) {
		schema := gin.H{
			"type":  "page",
			"title": "插件管理",
			"body": []any{
				gin.H{
					"type": "crud",
					"api":  "get:/admin/plugin/list",
					"columns": []any{
						gin.H{"name": "name", "label": "标识"},
						gin.H{"name": "title", "label": "名称"},
						gin.H{"name": "version", "label": "版本"},
						gin.H{"name": "description", "label": "描述"},
						gin.H{"name": "status", "label": "状态"},
					},
					"itemActions": []any{
						gin.H{
							"type":        "button",
							"actionType":  "ajax",
							"label":       "安装",
							"level":       "primary",
							"visibleOn":   "this.status == '已发现' || this.status == '已禁用'",
							"confirmText": "确认安装该插件？",
							"api":         "post:/admin/plugin/install/${name}",
						},
						gin.H{
							"type":        "button",
							"actionType":  "ajax",
							"label":       "卸载",
							"level":       "danger",
							"visibleOn":   "this.status != '已发现'",
							"confirmText": "确认卸载该插件？卸载后需要重新注册才可再次安装。",
							"api":         "post:/admin/plugin/uninstall/${name}",
						},
					},
				},
			},
		}
		amis.WriteJsonData(c, schema)
	})
}

// Register 注册插件模块，默认状态为已发现
func (m *Manager) Register(module Module) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	name := module.Meta.Name
	if name == "" {
		return fmt.Errorf("插件名称不能为空")
	}
	if _, ok := m.modules[name]; ok {
		return fmt.Errorf("插件已存在: %s", name)
	}
	m.modules[name] = module
	m.status[name] = StatusDiscovered
	klog.V(6).Infof("注册插件: %s（版本: %s）", module.Meta.Name, module.Meta.Version)
	return nil
}

// Install 安装指定插件（幂等），调用生命周期的 Install
func (m *Manager) Install(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	mod, ok := m.modules[name]
	if !ok {
		return fmt.Errorf("插件未注册: %s", name)
	}
	if mod.Lifecycle != nil {
		ctx := installContextImpl{baseContextImpl{meta: mod.Meta}}
		if err := mod.Lifecycle.Install(ctx); err != nil {
			klog.V(6).Infof("安装插件失败: %s，错误: %v", name, err)
			return err
		}
	}
	m.status[name] = StatusInstalled
	klog.V(6).Infof("安装插件成功: %s", name)
	return nil
}

// Enable 启用指定插件，调用生命周期的 Enable
func (m *Manager) Enable(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	mod, ok := m.modules[name]
	if !ok {
		return fmt.Errorf("插件未注册: %s", name)
	}
	if mod.Lifecycle != nil {
		ctx := enableContextImpl{baseContextImpl{meta: mod.Meta}}
		if err := mod.Lifecycle.Enable(ctx); err != nil {
			klog.V(6).Infof("启用插件失败: %s，错误: %v", name, err)
			return err
		}
	}
	m.status[name] = StatusEnabled
	klog.V(6).Infof("启用插件成功: %s", name)
	return nil
}

// Disable 禁用指定插件，调用生命周期的 Disable
func (m *Manager) Disable(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	mod, ok := m.modules[name]
	if !ok {
		return fmt.Errorf("插件未注册: %s", name)
	}
	if mod.Lifecycle != nil {
		ctx := baseContextImpl{meta: mod.Meta}
		if err := mod.Lifecycle.Disable(ctx); err != nil {
			klog.V(6).Infof("禁用插件失败: %s，错误: %v", name, err)
			return err
		}
	}
	m.status[name] = StatusDisabled
	klog.V(6).Infof("禁用插件成功: %s", name)
	return nil
}

// Uninstall 卸载指定插件（可选），调用生命周期的 Uninstall
func (m *Manager) Uninstall(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	mod, ok := m.modules[name]
	if !ok {
		return fmt.Errorf("插件未注册: %s", name)
	}
	if mod.Lifecycle != nil {
		ctx := installContextImpl{baseContextImpl{meta: mod.Meta}}
		if err := mod.Lifecycle.Uninstall(ctx); err != nil {
			klog.V(6).Infof("卸载插件失败: %s，错误: %v", name, err)
			return err
		}
	}
	delete(m.modules, name)
	delete(m.status, name)
	klog.V(6).Infof("卸载插件成功: %s", name)
	return nil
}

// IsEnabled 返回插件是否处于启用状态
func (m *Manager) IsEnabled(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.status[name] == StatusEnabled
}

// StatusOf 获取插件当前状态
func (m *Manager) StatusOf(name string) (Status, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, ok := m.status[name]
	return s, ok
}

// registrar 集中注册器，由外部包绑定，Start 时调用
var registrar func(*Manager)

// SetRegistrar 绑定集中注册器（仅设置，不做启停）
func SetRegistrar(f func(*Manager)) {
	registrar = f
}

// Start 启动插件管理：集中注册 + 默认启用策略
func (m *Manager) Start() {
	if registrar != nil {
		registrar(m)
	}
	m.EnableAll()
}

// EnableAll 启用所有已注册插件（用于默认策略或初始化）
func (m *Manager) EnableAll() {
	m.mu.Lock()
	defer m.mu.Unlock()
	for name, mod := range m.modules {
		if mod.Lifecycle != nil {
			ctx := enableContextImpl{baseContextImpl{meta: mod.Meta}}
			if err := mod.Lifecycle.Enable(ctx); err != nil {
				klog.V(6).Infof("启用插件失败: %s，错误: %v", name, err)
				continue
			}
		}
		m.status[name] = StatusEnabled
		klog.V(6).Infof("启用插件成功: %s", name)
	}
}

// RegisterRoutes 扫描已启用插件并注册其路由
func (m *Manager) RegisterRoutes(api *gin.RouterGroup) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for name, mod := range m.modules {
		if m.status[name] == StatusEnabled && mod.Router != nil {
			klog.V(6).Infof("注册插件路由: %s", name)
			mod.Router(api)
		}
	}
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

// baseContextImpl 基础上下文实现
type baseContextImpl struct {
	meta Meta
}

// Meta 返回插件元信息
func (c baseContextImpl) Meta() Meta { return c.meta }

// Logger 返回日志接口占位
func (c baseContextImpl) Logger() Logger { return nil }

// Config 返回插件配置占位
func (c baseContextImpl) Config() PluginConfig { return nil }

// installContextImpl 安装期上下文实现
type installContextImpl struct {
	baseContextImpl
}

// DB 返回安装期DB接口占位
func (c installContextImpl) DB() SchemaOperator { return nil }

// ConfigRegistry 返回配置注册接口占位
func (c installContextImpl) ConfigRegistry() ConfigRegistry { return nil }

// enableContextImpl 启用期上下文实现
type enableContextImpl struct {
	baseContextImpl
}

// MenuRegistry 返回菜单注册接口占位
func (c enableContextImpl) MenuRegistry() MenuRegistry { return nil }

// PermissionRegistry 返回权限注册接口占位
func (c enableContextImpl) PermissionRegistry() PermissionRegistry { return nil }

// PageRegistry 返回页面注册接口占位
func (c enableContextImpl) PageRegistry() AmisPageRegistry { return nil }
