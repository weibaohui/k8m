package plugins

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

// upgradeContextImpl 升级期上下文实现
type upgradeContextImpl struct {
	baseContextImpl
	from string
	to   string
}

// FromVersion 返回旧版本
func (c upgradeContextImpl) FromVersion() string { return c.from }

// ToVersion 返回新版本
func (c upgradeContextImpl) ToVersion() string { return c.to }

// DB 返回升级期迁移接口占位
func (c upgradeContextImpl) DB() MigrationOperator { return nil }
