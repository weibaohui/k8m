package plugins

// Meta 插件元信息，用于系统识别与展示
type Meta struct {
	// Name 插件唯一标识（系统级唯一）
	Name string
	// Title 插件展示名称
	Title string
	// Version 插件版本号
	Version string
	// Description 插件功能描述
	Description string
}

// Menu 菜单模型定义，仅用于声明结构
type Menu struct {
	// ID 菜单唯一标识
	ID string
	// Title 菜单展示标题
	Title string
	// Path 指向 AMIS 页面路径
	Path string
	// Permission 访问所需权限名称
	Permission string
}

// Permission 权限定义
type Permission struct {
	// Name 权限唯一名称（全局唯一）
	Name string
	// Title 权限展示名称
	Title string
}

// APIRoute 后端 API 路由声明（不含实现）
type APIRoute struct {
	// Method HTTP 方法，如 GET/POST/PUT/DELETE
	Method string
	// Path API 路径，需以 /api/plugins/<plugin-name>/ 开头
	Path string
	// RequiredPermission 访问该 API 所需权限名称（可选）
	RequiredPermission string
}

// FrontendResource 前端 AMIS JSON 资源声明
type FrontendResource struct {
	// Name 资源标识
	Name string
	// Path 资源路径，通常位于插件 frontend/ 目录下
	Path string
}

// Table 数据表结构声明（仅定义，不做具体实现）
type Table struct {
	// Name 表名称，建议包含插件名前缀
	Name string
	// Schema 表结构定义（例如 DDL 片段或结构化描述）
	Schema string
}

