package plugins

// Module 插件（Feature Module）声明体，仅用于描述能力集合
type Module struct {
	// Meta 插件元信息（系统识别与展示）
	Meta Meta
	// Menus 菜单声明（0..n）
	Menus []Menu
	// Permissions 权限声明（0..n）
	Permissions []Permission
	// APIRoutes 后端 API 路由声明（0..n）
	APIRoutes []APIRoute
	// Frontend 前端 AMIS JSON 资源声明（0..n）
	Frontend []FrontendResource
	// Tables 数据表结构声明（仅 Install/Uninstall 生效）
	Tables []Table
	// Lifecycle 生命周期实现（由系统调度调用）
	Lifecycle Lifecycle
}

