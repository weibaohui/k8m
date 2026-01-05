package plugins

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/robfig/cron/v3"
	"github.com/weibaohui/k8m/pkg/plugins/eventbus"
	"k8s.io/klog/v2"
)

// Manager 管理插件的注册、安装、启用和禁用
type Manager struct {
	mu            sync.RWMutex                       // 读写锁，保护插件并发安全
	modules       map[string]Module                  // 已注册的插件模块，key为模块名
	status        map[string]Status                  // 插件状态，key为模块名
	apiGroups     []chi.Router                       // 插件API路由组，用于注册HTTP路由
	engine        chi.Router                         // Chi路由引擎，用于构建HTTP路由
	cron          *cron.Cron                         // 定时任务调度器
	cronIDs       map[string]map[string]cron.EntryID // 定时任务ID，key为模块名+任务名
	cronRunning   map[string]map[string]bool         // 定时任务运行状态，key为模块名+任务名
	atomicHandler *AtomicHandler                     // 用于插件启用禁用后动态处理api路由。原子处理器，用于并发安全的处理器管理
	routerBuilder func(chi.Router) http.Handler      // 路由构建函数，用于构建最终HTTP处理器。用于插件启用禁用后动态处理api路由
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
		modules: make(map[string]Module),
		status:  make(map[string]Status),

		apiGroups: make([]chi.Router, 0),
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
	name := module.Meta.Name
	if name == "" {
		return fmt.Errorf("插件名称不能为空")
	}
	m.mu.RLock()
	_, ok := m.modules[name]
	m.mu.RUnlock()
	if ok {
		return fmt.Errorf("插件已存在: %s", name)
	}
	m.mu.Lock()
	m.modules[name] = module
	m.status[name] = StatusDiscovered
	m.mu.Unlock()
	klog.V(6).Infof("注册插件: %s（版本: %s）", module.Meta.Name, module.Meta.Version)
	return nil
}

// Install 安装指定插件（幂等），调用生命周期的 Install
// 注意：该方法用于实际启停周期调用，非管理员API配置写入
func (m *Manager) Install(name string) error {
	m.mu.RLock()
	mod, ok := m.modules[name]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("插件未注册: %s", name)
	}
	if mod.Lifecycle != nil {
		ctx := installContextImpl{baseContextImpl{meta: mod.Meta, bus: eventbus.New()}}
		if err := mod.Lifecycle.Install(ctx); err != nil {
			klog.V(6).Infof("安装插件失败: %s，错误: %v", name, err)
			return err
		}
	}
	m.mu.Lock()
	m.status[name] = StatusInstalled
	m.mu.Unlock()
	klog.V(6).Infof("安装插件成功: %s", name)
	return nil
}

// Upgrade 升级指定插件（版本变更触发），调用生命周期的 Upgrade
// 该方法不改变当前状态，仅执行安全迁移逻辑
func (m *Manager) Upgrade(name string, fromVersion string, toVersion string) error {
	m.mu.RLock()
	mod, ok := m.modules[name]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("插件未注册: %s", name)
	}
	if mod.Lifecycle != nil {
		ctx := upgradeContextImpl{
			baseContextImpl: baseContextImpl{meta: mod.Meta, bus: eventbus.New()},
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

// Enable 启用指定插件，调用生命周期的 Enable
// 注意：该方法用于实际启停周期调用，非管理员API配置写入
func (m *Manager) Enable(name string) error {
	m.mu.RLock()
	mod, ok := m.modules[name]
	m.mu.RUnlock()
	if !ok {

		return fmt.Errorf("插件未注册: %s", name)
	}
	// 依赖检查：启用前必须确保所有依赖插件均已启用
	if len(mod.Dependencies) > 0 {
		for _, dep := range mod.Dependencies {
			m.mu.RLock()
			ds := m.status[dep]
			m.mu.RUnlock()
			if ds != StatusEnabled {
				klog.V(6).Infof("启用插件失败: %s，依赖未启用: %s", name, dep)
				return fmt.Errorf("依赖插件未启用: %s", dep)
			}
		}
	}
	if mod.Lifecycle != nil {
		ctx := enableContextImpl{baseContextImpl{meta: mod.Meta, bus: eventbus.New()}}
		if err := mod.Lifecycle.Enable(ctx); err != nil {
			klog.V(6).Infof("启用插件失败: %s，错误: %v", name, err)
			return err
		}
	}
	m.mu.Lock()
	m.status[name] = StatusEnabled
	m.mu.Unlock()

	klog.V(6).Infof("启用插件成功: %s", name)
	m.rebuildRouter()

	return nil
}

// Disable 禁用指定插件,调用生命周期的 Disable
// 注意:该方法用于实际启停周期调用,非管理员API配置写入
func (m *Manager) Disable(name string) error {
	m.mu.RLock()
	mod, ok := m.modules[name]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("插件未注册: %s", name)
	}
	for otherName, otherMod := range m.modules {
		m.mu.RLock()
		ost := m.status[otherName]
		m.mu.RUnlock()
		if ost != StatusEnabled {
			continue
		}
		for _, dep := range otherMod.Dependencies {
			if dep == name {
				klog.V(6).Infof("禁用插件失败: %s,被插件依赖: %s", name, otherName)
				return fmt.Errorf("无法禁用插件,插件 %s 依赖于当前插件", otherName)
			}
		}
	}
	if mod.Lifecycle != nil {
		ctx := baseContextImpl{meta: mod.Meta, bus: eventbus.New()}
		if err := mod.Lifecycle.Disable(ctx); err != nil {
			klog.V(6).Infof("禁用插件失败: %s,错误: %v", name, err)
			return err
		}
	}
	m.mu.Lock()
	m.status[name] = StatusDisabled
	m.mu.Unlock()
	klog.V(6).Infof("禁用插件成功: %s", name)

	m.rebuildRouter()

	return nil
}

// Uninstall 卸载指定插件（可选），调用生命周期的 Uninstall
// 注意：该方法用于实际启停周期调用，非管理员API配置写入
func (m *Manager) Uninstall(name string, keepData bool) error {
	m.mu.RLock()
	mod, ok := m.modules[name]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("插件未注册: %s", name)
	}
	if mod.Lifecycle != nil {
		ctx := uninstallContextImpl{baseContextImpl{meta: mod.Meta, bus: eventbus.New()}, keepData}
		if err := mod.Lifecycle.Uninstall(ctx); err != nil {
			klog.V(6).Infof("卸载插件失败: %s，错误: %v", name, err)
			return err
		}
	}
	m.mu.Lock()
	m.status[name] = StatusDiscovered
	m.mu.Unlock()
	klog.V(6).Infof("卸载插件成功: %s", name)
	return nil
}

// IsEnabled 返回插件是否处于启用状态
func (m *Manager) IsEnabled(name string) bool {
	m.mu.RLock()
	enabled := m.status[name] == StatusEnabled
	m.mu.RUnlock()
	return enabled
}

// StatusOf 获取插件当前状态
func (m *Manager) StatusOf(name string) (Status, bool) {
	m.mu.RLock()
	s, ok := m.status[name]
	m.mu.RUnlock()
	return s, ok
}

// registrar 集中注册器，由外部包绑定，Start 时调用
var registrar func(*Manager)

// SetRegistrar 绑定集中注册器（仅设置，不做启停）
func SetRegistrar(f func(*Manager)) {
	registrar = f
}

// SetEngine 设置 Chi 引擎
// 便于后续统计与展示插件已注册的路由
func (m *Manager) SetEngine(e chi.Router) {
	m.mu.Lock()
	m.engine = e
	m.mu.Unlock()
}

// topologicalSort 对已启用的插件进行拓扑排序,确保依赖先启动
// 返回插件名称列表,按依赖顺序排列(被依赖的插件在前)
// 同时支持 RunAfter 字段,确保插件在 RunAfter 列表中的插件之后启动
func (m *Manager) topologicalSort() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 构建入度表和邻接表(仅针对已启用的插件)
	inDegree := make(map[string]int)
	graph := make(map[string][]string) // 反向图:dependency -> dependents
	enabledNames := make([]string, 0)

	// 初始化:收集所有已启用的插件
	for name, status := range m.status {
		if status == StatusEnabled {
			enabledNames = append(enabledNames, name)
			inDegree[name] = 0
		}
	}

	// 构建图和计算入度
	for _, name := range enabledNames {
		mod, ok := m.modules[name]
		if !ok {
			continue
		}
		// 处理 Dependencies:强依赖关系
		for _, dep := range mod.Dependencies {
			// 只处理已启用的依赖
			if m.status[dep] == StatusEnabled {
				graph[dep] = append(graph[dep], name)
				inDegree[name]++
			}
		}
		// 处理 RunAfter:启动顺序约束,必须在指定插件之后启动
		for _, runAfter := range mod.RunAfter {
			// 只处理已启用的插件
			if m.status[runAfter] == StatusEnabled {
				graph[runAfter] = append(graph[runAfter], name)
				inDegree[name]++
			}
		}
	}

	// 拓扑排序（Kahn算法）
	queue := make([]string, 0)
	for _, name := range enabledNames {
		if inDegree[name] == 0 {
			queue = append(queue, name)
		}
	}

	result := make([]string, 0, len(enabledNames))
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		result = append(result, current)

		for _, dependent := range graph[current] {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
			}
		}
	}

	// 检测循环依赖
	if len(result) != len(enabledNames) {
		klog.V(6).Infof("警告：检测到循环依赖，部分插件可能无法正确排序")
		// 将未排序的插件追加到结果末尾
		for _, name := range enabledNames {
			found := false
			for _, r := range result {
				if r == name {
					found = true
					break
				}
			}
			if !found {
				result = append(result, name)
			}
		}
	}

	return result
}

// Start 启动插件管理：集中注册 + 默认启用策略
func (m *Manager) Start() {
	if registrar != nil {
		registrar(m)
	}
	m.ApplyConfigFromDB()
	//增加插件启动任务管理
	//按依赖顺序逐个启动插件中注册的后台任务。不阻塞
	sortedNames := m.topologicalSort()
	for _, name := range sortedNames {
		mod, ok := m.modules[name]
		if !ok {
			continue
		}
		if mod.Lifecycle != nil && m.status[name] == StatusEnabled {
			ctx := baseContextImpl{meta: mod.Meta, bus: eventbus.New()}
			if err := mod.Lifecycle.Start(ctx); err != nil {
				klog.V(6).Infof("启动插件后台任务失败: %s，错误: %v", name, err)
			} else {
				klog.V(6).Infof("启动插件后台任务成功: %s", name)
			}
		}
	}

	//逐个启动插件中定义的cron表达式的定时任务
	// 同样按依赖顺序注册
	for _, name := range sortedNames {
		mod, ok := m.modules[name]
		if !ok {
			continue
		}
		if mod.Lifecycle == nil || len(mod.Crons) == 0 || m.status[name] != StatusEnabled {
			continue
		}
		ctx := baseContextImpl{meta: mod.Meta, bus: eventbus.New()}
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
	m.rebuildRouter()
}

func (m *Manager) rebuildRouter() {
	if m.atomicHandler == nil || m.routerBuilder == nil {
		klog.V(6).Infof("路由重建跳过: atomicHandler=%v, routerBuilder=%v", m.atomicHandler != nil, m.routerBuilder != nil)
		return
	}

	newRouter := chi.NewRouter()
	newHandler := m.routerBuilder(newRouter)
	m.atomicHandler.Store(newHandler)
	m.engine = newRouter
	klog.V(6).Infof("路由树已重新构建")
}

func (m *Manager) SetAtomicHandler(ah *AtomicHandler) {
	m.mu.Lock()
	m.atomicHandler = ah
	m.mu.Unlock()
	m.rebuildRouter()
}

func (m *Manager) SetRouterBuilder(builder func(chi.Router) http.Handler) {
	m.mu.Lock()
	m.routerBuilder = builder
	m.mu.Unlock()
}
