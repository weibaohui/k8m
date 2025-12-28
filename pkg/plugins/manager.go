package plugins

import (
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron/v3"
	"k8s.io/klog/v2"
)

// Manager 管理插件的注册、安装、启用和禁用
type Manager struct {
	mu      sync.RWMutex
	modules map[string]Module
	status  map[string]Status
	// apiGroups 已注册的后端API路由分组引用，用于支持启用时动态注册路由
	apiGroups []*gin.RouterGroup
	// engine Gin 引擎引用，用于统计已注册路由
	engine *gin.Engine
	// cron 定时任务调度器
	cron *cron.Cron
	// cronIDs 记录每个插件每条 cron 的 EntryID
	cronIDs map[string]map[string]cron.EntryID
	// cronRunning 记录每个插件每条 cron 是否正在运行
	cronRunning map[string]map[string]bool
}

var (
	managerOnce     sync.Once
	managerInstance *Manager
)

// ManagerInstance 返回全局唯一的插件管理器实例
func ManagerInstance() *Manager {
	managerOnce.Do(func() {
		managerInstance = newManager()
	})
	return managerInstance
}

// NewManager 创建并返回插件管理器（单例）
func NewManager() *Manager {
	return ManagerInstance()
}

// newManager 创建并返回插件管理器
func newManager() *Manager {
	return &Manager{
		modules:   make(map[string]Module),
		status:    make(map[string]Status),
		apiGroups: make([]*gin.RouterGroup, 0),
		cron: cron.New(
			cron.WithChain(
				cron.Recover(cron.DefaultLogger),
				cron.SkipIfStillRunning(cron.DefaultLogger),
			),
		),
		cronIDs:     make(map[string]map[string]cron.EntryID),
		cronRunning: make(map[string]map[string]bool),
	}
}

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
	// 依赖检查：启用前必须确保所有依赖插件均已启用
	if len(mod.Dependencies) > 0 {
		for _, dep := range mod.Dependencies {
			if m.status[dep] != StatusEnabled {
				klog.V(6).Infof("启用插件失败: %s，依赖未启用: %s", name, dep)
				return fmt.Errorf("依赖插件未启用: %s", dep)
			}
		}
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

// Upgrade 升级指定插件（版本变更触发），调用生命周期的 Upgrade
// 该方法不改变当前状态，仅执行安全迁移逻辑
func (m *Manager) Upgrade(name string, fromVersion string, toVersion string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	mod, ok := m.modules[name]
	if !ok {
		return fmt.Errorf("插件未注册: %s", name)
	}
	if mod.Lifecycle != nil {
		ctx := upgradeContextImpl{
			baseContextImpl: baseContextImpl{meta: mod.Meta},
			from:            fromVersion,
			to:              toVersion,
		}
		if err := mod.Lifecycle.Upgrade(ctx); err != nil {
			klog.V(6).Infof("升级插件失败: %s，从 %s 到 %s，错误: %v", name, fromVersion, toVersion, err)
			return err
		}
	}
	klog.V(6).Infof("升级插件成功: %s，从 %s 到 %s", name, fromVersion, toVersion)
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

// SetEngine 设置 Gin 引擎
// 便于后续统计与展示插件已注册的路由
func (m *Manager) SetEngine(e *gin.Engine) {
	m.mu.Lock()
	m.engine = e
	m.mu.Unlock()
}

// Start 启动插件管理：集中注册 + 默认启用策略
func (m *Manager) Start() {
	if registrar != nil {
		registrar(m)
	}
	m.ApplyConfigFromDB()
	//增加插件启动任务管理
	//逐个启动插件中注册的后台任务。不阻塞
	for name, mod := range m.modules {
		if mod.Lifecycle != nil && m.status[name] == StatusEnabled {
			ctx := baseContextImpl{meta: mod.Meta}
			if err := mod.Lifecycle.Start(ctx); err != nil {
				klog.V(6).Infof("启动插件后台任务失败: %s，错误: %v", name, err)
			} else {
				klog.V(6).Infof("启动插件后台任务成功: %s", name)
			}
		}
	}

	//逐个启动插件中定义的cron表达式的定时任务
	for name, mod := range m.modules {
		if mod.Lifecycle == nil || len(mod.Crons) == 0 || m.status[name] != StatusEnabled {
			continue
		}
		ctx := baseContextImpl{meta: mod.Meta}
		for _, spec := range mod.Crons {
			if _, err := cron.ParseStandard(spec); err != nil {
				klog.V(6).Infof("插件 %s 的 cron 表达式非法: %s，错误: %v", name, spec, err)
				continue
			}
			n := name
			s := spec
			id, err := m.cron.AddFunc(s, func() {
				m.setCronRunning(n, s, true)
				if err := mod.Lifecycle.StartCron(ctx, s); err != nil {
					klog.V(6).Infof("执行插件定时任务失败: %s，表达式: %s，错误: %v", n, s, err)
				} else {
					klog.V(6).Infof("执行插件定时任务成功: %s，表达式: %s", n, s)
				}
				m.setCronRunning(n, s, false)
			})
			if err != nil {
				klog.V(6).Infof("注册插件定时任务失败: %s，表达式: %s，错误: %v", name, s, err)
			} else {
				if _, ok := m.cronIDs[name]; !ok {
					m.cronIDs[name] = make(map[string]cron.EntryID)
				}
				m.cronIDs[name][s] = id
				klog.V(6).Infof("注册插件定时任务成功: %s，表达式: %s", name, s)
			}
		}
	}
	m.cron.Start()
}
