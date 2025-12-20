package plugins

import (
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/models"
	"k8s.io/klog/v2"
)

// Manager 管理插件的注册、安装、启用和禁用
type Manager struct {
	mu      sync.RWMutex
	modules map[string]Module
	status  map[string]Status
	// apiGroups 已注册的后端API路由分组引用，用于支持启用时动态注册路由
	apiGroups []*gin.RouterGroup
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
		modules:   make(map[string]Module),
		status:    make(map[string]Status),
		apiGroups: make([]*gin.RouterGroup, 0),
	}
}

// 管理相关API与页面Schema已迁移至 manager_api.go

// Register 注册插件模块，默认状态为已发现
// 仅在内存中标记为已发现，不进行任何持久化操作
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
// 注意：该方法用于实际启停周期调用，非管理员API配置写入
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
// 注意：该方法用于实际启停周期调用，非管理员API配置写入
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
// 注意：该方法用于实际启停周期调用，非管理员API配置写入
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
// 注意：该方法用于实际启停周期调用，非管理员API配置写入
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
	// 卸载后保留插件条目，使其仍然显示在列表中并可再次安装
	m.status[name] = StatusDiscovered
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
	m.ApplyConfigFromDB()
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

// RegisterRoutes 扫描已启用插件并注册其路由，并记录路由分组用于后续动态注册
func (m *Manager) RegisterRoutes(api *gin.RouterGroup) {
	m.mu.Lock()
	defer m.mu.Unlock()
	// 记录路由分组（去重）
	already := false
	for _, g := range m.apiGroups {
		if g == api {
			already = true
			break
		}
	}
	if !already {
		m.apiGroups = append(m.apiGroups, api)
	}
	// 为已启用插件注册路由
	for name, mod := range m.modules {
		if m.status[name] == StatusEnabled && mod.Router != nil {
			klog.V(6).Infof("注册插件路由: %s", name)
			mod.Router(api)
		}
	}
}

// PersistStatus 将插件状态持久化到数据库
// 管理员API调用该方法写入配置，实际生效需要重启
func (m *Manager) PersistStatus(name string, status Status, params *dao.Params) error {
	if _, ok := m.modules[name]; !ok {
		return fmt.Errorf("插件未注册: %s", name)
	}
	cfg := &models.PluginConfig{
		Name:   name,
		Status: statusToString(status),
	}
	return cfg.SaveByName(params)
}

// ApplyConfigFromDB 启动时从数据库加载插件配置并应用
// 根据持久化状态执行安装或启用操作；未配置的插件默认启用并写入数据库
func (m *Manager) ApplyConfigFromDB() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 读取所有持久化配置
	params := dao.BuildDefaultParams()
	cfgModel := &models.PluginConfig{}
	cfgs, _, err := cfgModel.List(params)
	if err != nil {
		klog.V(6).Infof("读取插件配置失败: %v", err)
	}
	cfgMap := make(map[string]Status, len(cfgs))
	for _, c := range cfgs {
		cfgMap[c.Name] = statusFromString(c.Status)
	}

	// 应用配置
	for name, mod := range m.modules {
		st, ok := cfgMap[name]
		if !ok {
			// 默认启用并写入数据库
			st = StatusEnabled
			_ = (&models.PluginConfig{Name: name, Status: statusToString(st)}).SaveByName(params)
		}
		m.status[name] = st

		switch st {
		case StatusInstalled:
			if mod.Lifecycle != nil {
				ctx := installContextImpl{baseContextImpl{meta: mod.Meta}}
				if err := mod.Lifecycle.Install(ctx); err != nil {
					klog.V(6).Infof("启动时安装插件失败: %s，错误: %v", name, err)
				} else {
					klog.V(6).Infof("启动时安装插件成功: %s", name)
				}
			}
		case StatusEnabled:
			if mod.Lifecycle != nil {
				ctx := enableContextImpl{baseContextImpl{meta: mod.Meta}}
				if err := mod.Lifecycle.Enable(ctx); err != nil {
					klog.V(6).Infof("启动时启用插件失败: %s，错误: %v", name, err)
				} else {
					klog.V(6).Infof("启动时启用插件成功: %s", name)
				}
			}
		case StatusDisabled:
			klog.V(6).Infof("启动时禁用插件: %s", name)
		case StatusDiscovered:
			klog.V(6).Infof("启动时标记插件为已发现: %s", name)
		}
	}
}

// statusToString 状态转字符串
func statusToString(s Status) string {
	switch s {
	case StatusDiscovered:
		return "discovered"
	case StatusInstalled:
		return "installed"
	case StatusEnabled:
		return "enabled"
	case StatusDisabled:
		return "disabled"
	default:
		return "unknown"
	}
}

// statusFromString 字符串转状态
func statusFromString(s string) Status {
	switch s {
	case "discovered":
		return StatusDiscovered
	case "installed":
		return StatusInstalled
	case "enabled":
		return StatusEnabled
	case "disabled":
		return StatusDisabled
	default:
		return StatusDiscovered
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
