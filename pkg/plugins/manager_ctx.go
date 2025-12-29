package plugins

// baseContextImpl 基础上下文实现
type baseContextImpl struct {
	meta Meta
}

// Meta 返回插件元信息
func (c baseContextImpl) Meta() Meta { return c.meta }

// installContextImpl 安装期上下文实现
type installContextImpl struct {
	baseContextImpl
}

// enableContextImpl 启用期上下文实现
type enableContextImpl struct {
	baseContextImpl
}

// upgradeContextImpl 升级期上下文实现
type upgradeContextImpl struct {
	baseContextImpl
	from string
	to   string
}

// uninstallContextImpl 卸载期上下文实现
type uninstallContextImpl struct {
	baseContextImpl
	keepData bool
}

// KeepData 返回是否保留数据的选项
func (c uninstallContextImpl) KeepData() bool { return c.keepData }

// FromVersion 返回旧版本
func (c upgradeContextImpl) FromVersion() string { return c.from }

// ToVersion 返回新版本
func (c upgradeContextImpl) ToVersion() string { return c.to }
