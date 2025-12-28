package plugins

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

// PluginItemVO 插件列表展示结构体
// 用于在管理员接口中返回插件的基础信息与当前状态
type PluginItemVO struct {
	Name         string   `json:"name"`
	Title        string   `json:"title"`
	Version      string   `json:"version"`
	DbVersion    string   `json:"dbVersion,omitempty"`
	CanUpgrade   bool     `json:"canUpgrade,omitempty"`
	Description  string   `json:"description"`
	Status       string   `json:"status"`
	Menus        []Menu   `json:"menus,omitempty"`
	MenuCount    int      `json:"menuCount,omitempty"`
	CronCount    int      `json:"cronCount,omitempty"`
	Dependencies []string `json:"dependencies,omitempty"`
}

// RegisterAdminRoutes 注册插件的管理员路由
// 管理员路由通常用于插件的配置、管理和操作接口，需要较高的权限才能访问。
// 提供功能：
// 1. 插件列表（显示Meta信息与状态）
// 2. 安装插件
// 3. 卸载插件
func (m *Manager) RegisterAdminRoutes(admin *gin.RouterGroup) {
	grp := admin.Group("/plugin")

	// 列出所有已注册插件的Meta和状态
	grp.GET("/list", m.ListPlugins)

	// 安装插件
	grp.POST("/install/:name", m.InstallPlugin)
	// 启用插件
	grp.POST("/enable/:name", m.EnablePlugin)
	// 禁用插件
	grp.POST("/disable/:name", m.DisablePlugin)
	// 卸载插件
	grp.POST("/uninstall/:name", m.UninstallPlugin)
	// 升级插件
	grp.POST("/upgrade/:name", m.UpgradePlugin)

	// 定时任务管理
	grp.GET("/cron/:name", m.ListPluginCrons)
	grp.POST("/cron/:name/run_once", m.RunPluginCronOnce)
	// 统一开关接口（生效/关闭）
	grp.POST("/cron/name/:name/spec/:spec/enabled/:enabled", m.SetPluginCronEnabled)

}

// RegisterParamRoutes 注册插件的参数路由
// 参数路由用于插件的参数配置接口，只要登录即可访问，类似公共参数。
// 提供功能：
// 1. 获取已启用插件的菜单数据
func (m *Manager) RegisterParamRoutes(params *gin.RouterGroup) {
	grp := params.Group("/plugin")
	// 获取已启用插件的菜单数据
	grp.GET("/menus", m.ListPluginMenus)
}

// ListPlugins 获取所有已注册插件的Meta与状态
// 返回插件名称、标题、版本、描述及当前状态（中文）
func (m *Manager) ListPlugins(c *gin.Context) {
	m.mu.RLock()
	defer m.mu.RUnlock()

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
			MenuCount:    len(mod.Menus),
			CronCount:    len(mod.Crons),
			Dependencies: mod.Dependencies,
		})
	}
	klog.V(6).Infof("获取插件列表，共计%d个", len(items))
	amis.WriteJsonListWithTotal(c, int64(len(items)), items)
}

// ListPluginMenus 获取所有已启用插件的菜单定义
// 返回前端可直接使用的菜单JSON（与前端 MenuItem 结构一致）
func (m *Manager) ListPluginMenus(c *gin.Context) {
	m.mu.RLock()
	defer m.mu.RUnlock()

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
		Permission  string   `json:"permission,omitempty"`
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
				Permission:  menu.Permission,
			})
		}
		return children
	}

	// 顶层按插件名称分组，将每个插件的菜单作为其子菜单
	var result []MenuVO
	for name, mod := range m.modules {
		if m.status[name] != StatusEnabled {
			continue
		}
		result = append(result, convertMenusToVO(mod.Menus)...)
	}
	klog.V(6).Infof("获取插件菜单列表，共计%d个", len(result))
	amis.WriteJsonData(c, result)
}

// InstallPlugin 安装指定名称的插件
// 路径参数为插件名，安装失败时返回错误
func (m *Manager) InstallPlugin(c *gin.Context) {
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

	amis.WriteJsonOKMsg(c, "已保存插件为已安装，需重启后生效")
}

// EnablePlugin 启用指定名称的插件
// 路径参数为插件名，启用失败时返回错误
func (m *Manager) EnablePlugin(c *gin.Context) {
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

	amis.WriteJsonOKMsg(c, "已保存插件为已启用，需重启后生效")
}

// UninstallPlugin 卸载指定名称的插件
// 路径参数为插件名，卸载失败时返回错误
func (m *Manager) UninstallPlugin(c *gin.Context) {
	name := c.Param("name")
	klog.V(6).Infof("卸载插件配置请求: %s", name)
	if err := m.Uninstall(name); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	params := dao.BuildParams(c)
	if err := m.PersistStatus(name, StatusDiscovered, params); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonOKMsg(c, "已保存插件为未安装（已发现），需重启后生效")
}

// DisablePlugin 禁用指定名称的插件
// 路径参数为插件名，禁用失败时返回错误
func (m *Manager) DisablePlugin(c *gin.Context) {
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

	amis.WriteJsonOKMsg(c, "已保存插件为已禁用，需重启后生效")
}

// UpgradePlugin 升级指定名称的插件（当代码版本高于数据库记录版本时）
// 路径参数为插件名，升级失败时返回错误
func (m *Manager) UpgradePlugin(c *gin.Context) {
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

	amis.WriteJsonOKMsg(c, "插件升级成功，已记录新版本，需重启后生效")
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
func (m *Manager) ListPluginCrons(c *gin.Context) {
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
func (m *Manager) StartPluginCron(c *gin.Context) {
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
func (m *Manager) EnablePluginCron(c *gin.Context) {
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
func (m *Manager) RunPluginCronOnce(c *gin.Context) {
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
	klog.V(6).Infof("手动执行插件定时任务一次成功: %s，表达式: %s", name, spec)
	amis.WriteJsonOK(c)
}

// StopPluginCron 强制停止（移除）指定插件的一条定时任务
func (m *Manager) StopPluginCron(c *gin.Context) {
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
func (m *Manager) SetPluginCronEnabled(c *gin.Context) {
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
