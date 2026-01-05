package plugins

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/response"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

// RegisterAdminRoutes 注册插件的管理员路由
// 管理员路由通常用于插件的配置、管理和操作接口，需要较高的权限才能访问。
// 提供功能：
// 1. 插件列表（显示Meta信息与状态）
// 2. 安装插件
// 3. 卸载插件

func (m *Manager) RegisterAdminRoutes(r chi.Router) {
	// 列出所有已注册插件的Meta和状态
	r.Get("/plugin/list", response.Adapter(m.ListPlugins))

	// 安装插件
	r.Post("/plugin/install/{name}", response.Adapter(m.InstallPlugin))
	// 启用插件
	r.Post("/plugin/enable/{name}", response.Adapter(m.EnablePlugin))
	// 禁用插件
	r.Post("/plugin/disable/{name}", response.Adapter(m.DisablePlugin))
	// 卸载插件（删除数据）
	r.Post("/plugin/uninstall/{name}", response.Adapter(m.UninstallPlugin))
	// 卸载插件（保留数据）
	r.Post("/plugin/uninstall-keep-data/{name}", response.Adapter(m.UninstallPluginKeepData))
	// 升级插件
	r.Post("/plugin/upgrade/{name}", response.Adapter(m.UpgradePlugin))

	// 定时任务管理
	r.Get("/plugin/cron/{name}", response.Adapter(m.ListPluginCrons))
	r.Post("/plugin/cron/{name}/run_once", response.Adapter(m.RunPluginCronOnce))
	// 统一开关接口（生效/关闭）
	r.Post("/plugin/cron/name/{name}/spec/{spec}/enabled/{enabled}", response.Adapter(m.SetPluginCronEnabled))

}

// RegisterParamRoutes 注册插件的参数路由
// 参数路由用于插件的参数配置接口，只要登录即可访问，类似公共参数。
// 提供功能：
// 1. 获取已启用插件的菜单数据

func (m *Manager) RegisterParamRoutes(r chi.Router) {
	// 获取已启用插件的菜单数据
	r.Get("/plugin/menus", response.Adapter(m.ListPluginMenus))

}

// countMenusRecursive 递归计算菜单总数，包括子菜单
func countMenusRecursive(menus []Menu) int {
	count := len(menus)
	for _, menu := range menus {
		count += countMenusRecursive(menu.Children)
	}
	return count
}

// ListPlugins 获取所有已注册插件的Meta与状态
// 返回插件名称、标题、版本、描述及当前状态（中文）
func (m *Manager) ListPlugins(c *response.Context) {

	items := make([]PluginItemVO, 0, len(m.modules))
	// 读取数据库中的配置状态
	params := dao.BuildDefaultParams()
	cfgs, _, _ := (&models.PluginConfig{}).List(params)
	cfgMap := make(map[string]string, len(cfgs))
	cfgVerMap := make(map[string]string, len(cfgs))
	for _, cfg := range cfgs {
		cfgMap[cfg.Name] = cfg.Status
		cfgVerMap[cfg.Name] = cfg.Version
	}
	for name, mod := range m.modules {
		// 优先使用数据库中的配置状态；若不存在则显示为已发现
		statusStr, ok := cfgMap[name]
		if !ok {
			statusStr = "discovered"
		}
		status := statusFromString(statusStr)
		dbVer := cfgVerMap[name]
		canUpgrade := statusStr != "discovered" && utils.CompareVersions(mod.Meta.Version, dbVer)
		items = append(items, PluginItemVO{
			Name:         mod.Meta.Name,
			Title:        mod.Meta.Title,
			Version:      mod.Meta.Version,
			DbVersion:    dbVer,
			CanUpgrade:   canUpgrade,
			Description:  mod.Meta.Description,
			Status:       statusToCN(status),
			Menus:        mod.Menus,
			MenuCount:    countMenusRecursive(mod.Menus),
			CronCount:    len(mod.Crons),
			Tables:       mod.Tables,
			TableCount:   len(mod.Tables),
			Dependencies: mod.Dependencies,
			RunAfter:     mod.RunAfter,
			Routes:       m.collectPluginRouteCategories(mod.Meta.Name),
		})
	}
	klog.V(8).Infof("获取插件列表，共计%d个", len(items))
	//对items 进行排序
	sort.Slice(items, func(i, j int) bool {
		return items[i].Name < items[j].Name
	})
	amis.WriteJsonListWithTotal(c, int64(len(items)), items)
}

// ListPluginMenus 获取所有已启用插件的菜单定义
// 返回前端可直接使用的菜单JSON（与前端 MenuItem 结构一致）
func (m *Manager) ListPluginMenus(c *response.Context) {

	type MenuVO struct {
		Key         string   `json:"key,omitempty"`
		Title       string   `json:"title"`
		Icon        string   `json:"icon,omitempty"`
		URL         string   `json:"url,omitempty"`
		EventType   string   `json:"eventType,omitempty"`
		CustomEvent string   `json:"customEvent,omitempty"`
		Order       float64  `json:"order,omitempty"`
		Children    []MenuVO `json:"children,omitempty"`
		Show        string   `json:"show,omitempty"`
	}

	// 递归转换插件菜单为前端结构
	var convertMenusToVO func([]Menu) []MenuVO
	convertMenusToVO = func(ms []Menu) []MenuVO {
		children := make([]MenuVO, 0, len(ms))
		for _, menu := range ms {
			children = append(children, MenuVO{
				Key:         menu.Key,
				Title:       menu.Title,
				Icon:        menu.Icon,
				URL:         menu.URL,
				EventType:   menu.EventType,
				CustomEvent: menu.CustomEvent,
				Order:       menu.Order,
				Children:    convertMenusToVO(menu.Children),
				Show:        menu.Show,
			})
		}
		return children
	}

	// 顶层按插件名称分组，将每个插件的菜单作为其子菜单
	var result []MenuVO
	for name, mod := range m.modules {

		m.mu.RLock()
		sn := m.status[name]
		m.mu.RUnlock()

		if sn != StatusEnabled {
			continue
		}
		result = append(result, convertMenusToVO(mod.Menus)...)
	}
	klog.V(6).Infof("获取插件菜单列表，共计%d个", len(result))
	amis.WriteJsonData(c, result)
}

// collectPluginRouteCategories 收集指定插件的三类路由数组
// 仅统计路径中包含 /plugins/{name}/ 的路由，并按类别归集
// 使用 chi.Walk 遍历路由并解析插件名和类别
func (m *Manager) collectPluginRouteCategories(name string) RouteCategoryVO {
	result := RouteCategoryVO{
		Cluster: []RouteItem{},
		Mgm:     []RouteItem{},
		Admin:   []RouteItem{},
	}

	if m.engine == nil {
		klog.V(6).Infof("路由引擎为空，无法收集路由信息")
		return result
	}

	chi.Walk(m.engine, func(
		method string,
		route string,
		handler http.Handler,
		middlewares ...func(http.Handler) http.Handler,
	) error {
		plugin, scope := parsePluginRoute(route)
		if plugin == "" || plugin != name {
			return nil
		}

		item := RouteItem{
			Method: method,
			Path:   route,
		}

		switch scope {
		case "mgm":
			result.Mgm = append(result.Mgm, item)
		case "k8s":
			result.Cluster = append(result.Cluster, item)
		case "admin":
			result.Admin = append(result.Admin, item)
		}

		return nil
	})

	return result
}

// parsePluginRoute 解析路由路径，提取插件名和类别
// 返回: 插件名, 类别(mgm/k8s/admin)
// 示例:
// - /mgm/plugins/demo/status -> demo, mgm
// - /k8s/cluster/{clusterID}/plugins/demo/pods -> demo, k8s
// - /admin/plugins/demo/reload -> demo, admin
func parsePluginRoute(route string) (string, string) {
	// /mgm/plugins/xxxx/...
	if strings.HasPrefix(route, "/mgm/plugins/") {
		rest := strings.TrimPrefix(route, "/mgm/plugins/")
		return firstSegment(rest), "mgm"
	}

	// /admin/plugins/xxxx/...
	if strings.HasPrefix(route, "/admin/plugins/") {
		rest := strings.TrimPrefix(route, "/admin/plugins/")
		return firstSegment(rest), "admin"
	}

	// /k8s/cluster/{clusterID}/plugins/xxxx/...
	if strings.HasPrefix(route, "/k8s/cluster/") &&
		strings.Contains(route, "/plugins/") {

		idx := strings.Index(route, "/plugins/")
		if idx > 0 {
			rest := route[idx+len("/plugins/"):]
			return firstSegment(rest), "k8s"
		}
	}

	return "", ""
}

// firstSegment 提取路径的第一段
// 示例: "demo/status" -> "demo"
func firstSegment(path string) string {
	if path == "" {
		return ""
	}
	parts := strings.SplitN(path, "/", 2)
	return parts[0]
}

// InstallPlugin 安装指定名称的插件
// 路径参数为插件名，安装失败时返回错误
func (m *Manager) InstallPlugin(c *response.Context) {
	name := c.Param("name")
	klog.V(6).Infof("安装插件配置请求: %s", name)

	if err := m.Install(name); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	params := dao.BuildParams(c)
	if err := m.PersistStatus(name, StatusInstalled, params); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonOKMsg(c, "已安装")
}

// EnablePlugin 启用指定名称的插件
// 路径参数为插件名，启用失败时返回错误
func (m *Manager) EnablePlugin(c *response.Context) {
	name := c.Param("name")
	klog.V(6).Infof("启用插件配置请求: %s", name)
	if err := m.Enable(name); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	params := dao.BuildParams(c)
	if err := m.PersistStatus(name, StatusEnabled, params); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonOKMsg(c, "已启用")
}

// UninstallPlugin 卸载指定名称的插件（删除数据）
// 路径参数为插件名，卸载失败时返回错误
func (m *Manager) UninstallPlugin(c *response.Context) {
	name := c.Param("name")
	klog.V(6).Infof("卸载插件配置请求(删除数据): %s", name)
	if err := m.Uninstall(name, false); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	params := dao.BuildParams(c)
	if err := m.PersistStatus(name, StatusDiscovered, params); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonOKMsg(c, "已卸载并删除数据")
}

// UninstallPluginKeepData 卸载指定名称的插件（保留数据）
// 路径参数为插件名，卸载失败时返回错误
func (m *Manager) UninstallPluginKeepData(c *response.Context) {
	name := c.Param("name")
	klog.V(6).Infof("卸载插件配置请求(保留数据): %s", name)
	if err := m.Uninstall(name, true); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	params := dao.BuildParams(c)
	if err := m.PersistStatus(name, StatusDiscovered, params); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonOKMsg(c, "已卸载但保留数据")
}

// DisablePlugin 禁用指定名称的插件
// 路径参数为插件名，禁用失败时返回错误
func (m *Manager) DisablePlugin(c *response.Context) {
	name := c.Param("name")
	klog.V(6).Infof("禁用插件配置请求: %s", name)
	if err := m.Disable(name); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	params := dao.BuildParams(c)
	if err := m.PersistStatus(name, StatusDisabled, params); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonOKMsg(c, "已禁用")
}

// UpgradePlugin 升级指定名称的插件（当代码版本高于数据库记录版本时）
// 路径参数为插件名，升级失败时返回错误
func (m *Manager) UpgradePlugin(c *response.Context) {
	name := c.Param("name")
	klog.V(6).Infof("升级插件配置请求: %s", name)

	// 获取模块与当前版本
	m.mu.RLock()
	mod, ok := m.modules[name]
	m.mu.RUnlock()
	if !ok {
		amis.WriteJsonError(c, fmt.Errorf("插件未注册: %s", name))
		return
	}
	toVersion := mod.Meta.Version

	// 从数据库读取当前记录版本
	params := dao.BuildParams(c)
	cfg, err := (&models.PluginConfig{}).GetOne(params, func(db *gorm.DB) *gorm.DB {
		return db.Where("name = ?", name)
	})
	if err != nil && err != gorm.ErrRecordNotFound {
		amis.WriteJsonError(c, err)
		return
	}
	fromVersion := ""
	if cfg != nil {
		fromVersion = cfg.Version
	}

	// 比较版本，只有代码版本大于数据库版本才允许升级
	if !utils.CompareVersions(toVersion, fromVersion) {
		amis.WriteJsonOKMsg(c, "无需升级，版本相同或更低")
		return
	}

	if err := m.Upgrade(name, fromVersion, toVersion); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 持久化当前状态与新版本
	// 保持现有状态不变，仅更新版本字段
	st, _ := m.StatusOf(name)
	if err := m.PersistStatus(name, st, params); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonOKMsg(c, "插件升级成功")
}

// CronItemVO 定时任务状态展示结构体
type CronItemVO struct {
	Spec       string `json:"spec"`
	Registered bool   `json:"registered"`
	Running    bool   `json:"running"`
	Next       string `json:"next,omitempty"`
	Prev       string `json:"prev,omitempty"`
}

// ListPluginCrons 获取指定插件的定时任务定义与状态
func (m *Manager) ListPluginCrons(c *response.Context) {
	name := c.Param("name")
	m.mu.RLock()
	mod, ok := m.modules[name]
	m.mu.RUnlock()
	if !ok {
		amis.WriteJsonError(c, fmt.Errorf("插件未注册: %s", name))
		return
	}
	items := make([]CronItemVO, 0, len(mod.Crons))
	for _, spec := range mod.Crons {
		entry, exists := m.getCronEntry(name, spec)
		running := false
		m.mu.RLock()
		if rm, ok := m.cronRunning[name]; ok {
			running = rm[spec]
		}
		m.mu.RUnlock()
		var nextStr, prevStr string
		if exists {
			if !entry.Next.IsZero() {
				nextStr = entry.Next.Format("2006-01-02 15:04:05")
			}
			if !entry.Prev.IsZero() {
				prevStr = entry.Prev.Format("2006-01-02 15:04:05")
			}
		}
		items = append(items, CronItemVO{
			Spec:       spec,
			Registered: exists,
			Running:    running,
			Next:       nextStr,
			Prev:       prevStr,
		})
	}
	klog.V(6).Infof("获取插件定时任务列表: %s，共计%d个", name, len(items))
	amis.WriteJsonListWithTotal(c, int64(len(items)), items)
}

// StartPluginCron 手动启动（注册）指定插件的一条定时任务
func (m *Manager) StartPluginCron(c *response.Context) {
	name := c.Param("name")
	spec := c.Query("spec")
	if spec == "" {
		var body struct {
			Spec string `json:"spec"`
		}
		_ = c.ShouldBindJSON(&body)
		spec = body.Spec
	}
	if spec == "" {
		amis.WriteJsonError(c, fmt.Errorf("缺少参数: spec"))
		return
	}
	if err := m.EnsureCron(name, spec); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	klog.V(6).Infof("生效插件定时任务成功: %s，表达式: %s", name, spec)
	amis.WriteJsonOK(c)
}

// EnablePluginCron 生效指定插件的一条定时任务（别名）
func (m *Manager) EnablePluginCron(c *response.Context) {
	name := c.Param("name")
	spec := c.Query("spec")
	if spec == "" {
		var body struct {
			Spec string `json:"spec"`
		}
		_ = c.ShouldBindJSON(&body)
		spec = body.Spec
	}
	if spec == "" {
		amis.WriteJsonError(c, fmt.Errorf("缺少参数: spec"))
		return
	}
	if err := m.EnsureCron(name, spec); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	klog.V(6).Infof("生效插件定时任务成功: %s，表达式: %s", name, spec)
	amis.WriteJsonOK(c)
}

// RunPluginCronOnce 立即执行指定插件的一条定时任务一次
func (m *Manager) RunPluginCronOnce(c *response.Context) {
	name := c.Param("name")
	spec := c.Query("spec")
	if spec == "" {
		var body struct {
			Spec string `json:"spec"`
		}
		_ = c.ShouldBindJSON(&body)
		spec = body.Spec
	}
	if spec == "" {
		amis.WriteJsonError(c, fmt.Errorf("缺少参数: spec"))
		return
	}
	if err := m.RunCronOnce(name, spec); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	klog.V(6).Infof("手动执行插件定时任务成功: %s，表达式: %s", name, spec)
	amis.WriteJsonOK(c)
}

// StopPluginCron 强制停止（移除）指定插件的一条定时任务
func (m *Manager) StopPluginCron(c *response.Context) {
	name := c.Param("name")
	spec := c.Query("spec")
	if spec == "" {
		var body struct {
			Spec string `json:"spec"`
		}
		_ = c.ShouldBindJSON(&body)
		spec = body.Spec
	}
	if spec == "" {
		amis.WriteJsonError(c, fmt.Errorf("缺少参数: spec"))
		return
	}
	m.RemoveCron(name, spec)
	klog.V(6).Infof("关闭插件定时任务成功: %s，表达式: %s", name, spec)
	amis.WriteJsonOK(c)
}

// SetPluginCronEnabled 设置插件定时任务开关（生效/关闭）
// 路径参数：name 插件名、spec cron 表达式、enabled true/false
// 行为：enabled=true 则生效（注册并调度）；enabled=false 则关闭（移除调度）
func (m *Manager) SetPluginCronEnabled(c *response.Context) {
	name := c.Param("name")
	spec := c.Param("spec")
	enabled := c.Param("enabled")
	if name == "" || spec == "" || enabled == "" {
		amis.WriteJsonError(c, fmt.Errorf("缺少参数: name/spec/enabled"))
		return
	}
	spec, err := utils.UrlSafeBase64Decode(spec)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	if enabled == "true" || enabled == "1" || enabled == "yes" {
		if err := m.EnsureCron(name, spec); err != nil {
			amis.WriteJsonError(c, err)
			return
		}
		klog.V(6).Infof("生效插件定时任务成功: %s，表达式: %s", name, spec)
		amis.WriteJsonOK(c)
		return
	}
	// 其他情况视为关闭
	m.RemoveCron(name, spec)
	klog.V(6).Infof("关闭插件定时任务成功: %s，表达式: %s", name, spec)
	amis.WriteJsonOK(c)
}
