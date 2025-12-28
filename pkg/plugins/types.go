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
	// 表达式中可使用的全局函数：
	// - isPlatformAdmin()：判断是否为平台管理员
	// - isUserHasRole('role')：判断用户是否有指定角色（role为字符串） guest platform_admin 两种
	// - isUserInGroup('group')：判断用户是否在指定组（group为字符串） 自定义的各种用户组名称
	// 返回值：true/false，用于判断是否显示该菜单
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

// RouteAccessKind 路由访问控制类型
// any：任何已登录用户可访问
// platform_admin：仅平台管理员可访问
// roles：指定角色列表可访问（平台管理员也视为拥有所有角色）
type RouteAccessKind string

const (
	AccessAnyUser       RouteAccessKind = "any"
	AccessPlatformAdmin RouteAccessKind = "platform_admin"
	AccessRoles         RouteAccessKind = "roles"
)

// RouteRule 路由访问控制规则
// 用于描述某个HTTP方法+路径的访问要求
type RouteRule struct {
	// Method HTTP方法，如 GET/POST
	Method string `json:"method"`
	// Path 路径，可写相对路径（不含插件前缀）或完整路径
	// 示例："/items" 或 "/plugins/demo/items"
	Path string `json:"path"`
	// Kind 访问控制类型
	Kind RouteAccessKind `json:"kind"`
	// Roles 指定角色列表（当 Kind=roles 时生效）
	Roles []string `json:"roles,omitempty"`
}
