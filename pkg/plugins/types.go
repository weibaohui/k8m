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

// Menu 菜单模型定义
// 该结构与前端 `MenuItem` 类型保持一致，支持动态融合到前端菜单中
// 字段说明：
// - key：菜单唯一标识
// - title：菜单展示标题
// - icon：Font Awesome 图标（可选）
// - url：URL跳转地址（可选，通常与 eventType=url 搭配）
// - eventType：事件类型，可选值 'url' | 'custom'
// - customEvent：自定义事件字符串，如 '() => loadJsonPage(\"/path\")'
// - order：排序号，支持小数（如 6.5）
// - children：子菜单列表
// - show：显示表达式，字符串形式的 JS 表达式
// - permission：访问所需权限名称（可选）
type Menu struct {
	// Key 菜单唯一标识
	Key string `json:"key,omitempty"`
	// Title 菜单展示标题
	Title string `json:"title"`
	// Icon 图标（Font Awesome 类名）
	Icon string `json:"icon,omitempty"`
	// URL 跳转地址
	URL string `json:"url,omitempty"`
	// EventType 事件类型：'url' 或 'custom'
	EventType string `json:"eventType,omitempty"`
	// CustomEvent 自定义事件，如：'() => loadJsonPage("/path")'
	CustomEvent string `json:"customEvent,omitempty"`
	// Order 排序号
	Order float64 `json:"order,omitempty"`
	// Children 子菜单
	Children []Menu `json:"children,omitempty"`
	// Show 显示表达式（字符串形式的JS表达式）
	Show string `json:"show,omitempty"`
	// Permission 访问所需权限名称
	Permission string `json:"permission,omitempty"`
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
