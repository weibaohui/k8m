package plugins

// Lifecycle 插件生命周期接口，禁止隐式行为
type Lifecycle interface {
	// Install 安装阶段；只执行一次，必须保证幂等；用于注册权限、创建表、初始化数据
	Install(ctx InstallContext) error
	// Upgrade 升级阶段；当版本变化时触发，用于安全迁移（SQL、数据、权限）
	Upgrade(ctx UpgradeContext) error
	// Enable 启用阶段；暴露运行期能力，如菜单、权限、AMIS 页面
	Enable(ctx EnableContext) error
	// Disable 禁用阶段；能力收敛，如隐藏菜单、撤销页面可访问（不删数据/权限）
	Disable(ctx BaseContext) error
	// Uninstall 卸载阶段（可选）；清理插件资源（如允许可删除表与初始化数据）
	Uninstall(ctx InstallContext) error
}
