package plugins

import "github.com/gin-gonic/gin"

// Module 插件（Feature Module）声明体，仅用于描述能力集合
type Module struct {
	// Meta 插件元信息（系统识别与展示）
	Meta Meta
	// Menus 菜单声明（0..n）
	Menus []Menu

	// Lifecycle 生命周期实现（由系统调度调用）
	Lifecycle Lifecycle
	// Router 路由注册回调（启用后由Manager统一挂载）
	Router func(api *gin.RouterGroup)
}
