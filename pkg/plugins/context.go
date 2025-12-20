package plugins

// Logger 日志接口（仅声明，不暴露具体实现）
type Logger interface{}

// PluginConfig 插件私有配置读取接口（仅声明）
type PluginConfig interface{}

// SchemaOperator 安装期的表结构操作接口（仅声明）
type SchemaOperator interface{}

// ConfigRegistry 插件级配置项注册接口（仅声明）
type ConfigRegistry interface{}

// MigrationOperator 升级期的迁移操作接口（仅声明）
type MigrationOperator interface{}

// MenuRegistry 菜单注册接口（仅声明）
type MenuRegistry interface{}

// PermissionRegistry 权限注册接口（仅声明）
type PermissionRegistry interface{}

// AmisPageRegistry AMIS 页面注册接口（仅声明）
type AmisPageRegistry interface{}

// QueryExecutor 运行期的数据查询接口（仅声明）
type QueryExecutor interface{}

// Scheduler 后台/定时任务调度接口（仅声明）
type Scheduler interface{}

// EventBus 事件总线接口（仅声明）
type EventBus interface{}

// BaseContext 基础上下文（所有生命周期共享，只读能力）
type BaseContext interface {
	// Meta 返回插件元信息（名称、版本）
	Meta() Meta
	// Logger 返回日志接口（自动携带插件名与生命周期）
	Logger() Logger
	// Config 返回插件私有配置读取入口
	Config() PluginConfig
}

// InstallContext 安装期上下文（只在首次安装时执行）
type InstallContext interface {
	BaseContext
	// DB 返回表结构操作接口，用于创建/删除表与初始化数据
	DB() SchemaOperator
	// ConfigRegistry 返回配置项注册接口，用于注册插件级配置
	ConfigRegistry() ConfigRegistry
}

// UpgradeContext 升级期上下文（版本变更触发，用于安全迁移）
type UpgradeContext interface {
	BaseContext
	// FromVersion 返回旧版本号
	FromVersion() string
	// ToVersion 返回新版本号
	ToVersion() string
	// DB 返回迁移操作接口，用于 SQL/数据/权限迁移
	DB() MigrationOperator
}

// EnableContext 启用期上下文（对外暴露能力的唯一阶段）
type EnableContext interface {
	BaseContext
	// MenuRegistry 返回菜单注册接口
	MenuRegistry() MenuRegistry
	// PermissionRegistry 返回权限注册接口
	PermissionRegistry() PermissionRegistry
	// PageRegistry 返回 AMIS 页面注册接口
	PageRegistry() AmisPageRegistry
}

// RunContext 运行期上下文（正常运行阶段，支撑任务与事件）
type RunContext interface {
	BaseContext
	// DB 返回数据查询接口（只读或受限写）
	DB() QueryExecutor
	// Scheduler 返回任务调度接口（后台任务/定时任务）
	Scheduler() Scheduler
	// EventBus 返回事件总线接口（事件监听与转发）
	EventBus() EventBus
}

