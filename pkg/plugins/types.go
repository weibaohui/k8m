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
	// 字符串true/false，是否显示该菜单
	// 注意：这是菜单的显示权限。
	// 后端API业务逻辑需调用plugins.EnsureUserIsPlatformAdmin(*gin.Context)等方法进行显式权限校验，
	// 后端API的权限校验不能依赖此表达式。
	Show string `json:"show,omitempty"`
}

// PluginItemVO 插件列表展示结构体
// 用于在管理员接口中返回插件的基础信息与当前状态
type PluginItemVO struct {
	Name         string          `json:"name"`
	Title        string          `json:"title"`
	Version      string          `json:"version"`
	DbVersion    string          `json:"dbVersion,omitempty"`
	CanUpgrade   bool            `json:"canUpgrade,omitempty"`
	Description  string          `json:"description"`
	Status       string          `json:"status"`
	Menus        []Menu          `json:"menus,omitempty"`
	MenuCount    int             `json:"menuCount,omitempty"`
	CronCount    int             `json:"cronCount,omitempty"`
	Dependencies []string        `json:"dependencies,omitempty"`
	Routes       RouteCategoryVO `json:"routes,omitempty"`
}

// RouteCategoryVO 路由类别
// 类别为 cluster/mgm/admin，routes 为该类别下的路由列表
type RouteCategoryVO struct {
	Cluster []RouteItem `json:"cluster,omitempty"`
	Admin   []RouteItem `json:"admin,omitempty"`
	Mgm     []RouteItem `json:"mgm,omitempty"`
}

// RouteItem 路由条目
// 展示 HTTP 方法、路径、处理器名
type RouteItem struct {
	Method  string `json:"method"`
	Path    string `json:"path"`
	Handler string `json:"handler,omitempty"`
}
