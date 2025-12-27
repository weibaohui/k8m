package plugins

import "github.com/gin-gonic/gin"

// Module 插件（Feature Module）声明体，仅用于描述能力集合
type Module struct {
	// Meta 插件元信息（系统识别与展示）
	Meta Meta
	// Menus 菜单声明（0..n）
	Menus []Menu
	// Dependencies 插件依赖的其他插件名称列表；启用前需确保均已启用
	Dependencies []string

	// Lifecycle 生命周期实现（由系统调度调用）
	Lifecycle Lifecycle
	// Crons 插件的定时任务调度表达式（5段 cron）
	Crons []string
	// ClusterRouter 路由注册回调（启用后由Manager统一挂载）
	// 该类API接口以/k8s/cluster/clusterID/plugins/xxx的形式暴露，带有集群ID
	// 通常是集群相关的操作页面，要求必须是已登录用户
	ClusterRouter func(cluster *gin.RouterGroup)
	// ManagementRouter 管理操作路由注册回调（启用后由Manager统一挂载）
	// 该类API接口以/mgm/plugins/xxx的形式暴露，不带集群ID
	// 通常是管理类的操作页面，要求必须是已登录用户
	ManagementRouter func(mgm *gin.RouterGroup)

	// PluginAdminRouter 插件管理员操作路由注册回调（启用后由Manager统一挂载）
	// 该类API接口以/admin/plugins/xxx的形式暴露，不带集群ID
	// 通常是插件管理员相关的操作页面,要求必须是平台管理员
	PluginAdminRouter func(admin *gin.RouterGroup)
}
