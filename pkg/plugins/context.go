package plugins

// BaseContext 基础上下文（所有生命周期共享，只读能力）
type BaseContext interface {
	// Meta 返回插件元信息（名称、版本）
	Meta() Meta
}

// InstallContext 安装期上下文（只在首次安装时执行）
type InstallContext interface {
	BaseContext
}

// UpgradeContext 升级期上下文（版本变更触发，用于安全迁移）
type UpgradeContext interface {
	BaseContext
	// FromVersion 返回旧版本号
	FromVersion() string
	// ToVersion 返回新版本号
	ToVersion() string
}

// EnableContext 启用期上下文（对外暴露能力的唯一阶段）
type EnableContext interface {
	BaseContext
}

// RunContext 运行期上下文（正常运行阶段，支撑任务与事件）
type RunContext interface {
	BaseContext
}
