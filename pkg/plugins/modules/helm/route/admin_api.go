package route

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/helm/admin"
	"k8s.io/klog/v2"
)

// RegisterPluginAdminRoutes 注册 Helm 插件的管理员路由（平台管理员）
// 对应原来的 /admin 路径下的路由
func RegisterPluginAdminRoutes(arg *gin.RouterGroup) {
	g := arg.Group("/plugins/" + modules.PluginNameHelm)
	ctrl := &admin.RepoController{}
	g.GET("/repo/list", ctrl.List)
	g.POST("/repo/delete/:ids", ctrl.Delete)
	g.POST("/repo/update_index", ctrl.UpdateReposIndex)
	g.POST("/repo/save", ctrl.Save)

	settingCtrl := &admin.SettingController{}
	g.GET("/setting/get", settingCtrl.GetSetting)
	g.POST("/setting/update", settingCtrl.UpdateSetting)

	klog.V(6).Infof("注册 Helm 插件管理路由(admin)")
}

// RegisterPluginAPIRoutes 注册 Helm 插件的 API 路由（K8s API）
// 对应原来的 /k8s/cluster/{cluster} 路径下的路由
func RegisterPluginAPIRoutes(arg *gin.RouterGroup) {
	api := arg.Group("/plugins/" + modules.PluginNameHelm)
	ctrl := &admin.ReleaseController{}

	api.GET("/release/list", ctrl.ListRelease)
	api.GET("/release/ns/:ns/name/:name/history/list", ctrl.ListReleaseHistory)
	api.POST("/release/:release/repo/:repo/chart/:chart/version/:version/install", ctrl.InstallRelease)
	api.POST("/release/ns/:ns/name/:name/uninstall", ctrl.UninstallRelease)
	api.GET("/release/ns/:ns/name/:name/revision/:revision/values", ctrl.GetReleaseValues)
	api.GET("/release/ns/:ns/name/:name/revision/:revision/notes", ctrl.GetReleaseNote)
	api.GET("/release/ns/:ns/name/:name/revision/:revision/install_log", ctrl.GetReleaseInstallLog)
	api.POST("/release/batch/uninstall", ctrl.BatchUninstallRelease)
	api.POST("/release/upgrade", ctrl.UpgradeRelease)

	klog.V(6).Infof("注册 Helm 插件 API 路由(api)")
}

// RegisterPluginMgmRoutes 注册 Helm 插件的管理路由（Mgm）
// 对应原来的 /mgm 路径下的路由
func RegisterPluginMgmRoutes(arg *gin.RouterGroup) {
	mgm := arg.Group("/plugins/" + modules.PluginNameHelm)
	ctrl := &admin.ChartController{}
	mgm.GET("/repo/:repo/chart/:chart/versions", ctrl.ChartVersionOptionList)
	mgm.GET("/repo/:repo/chart/:chart/version/:version/values", ctrl.GetChartValue)
	mgm.GET("/chart/list", ctrl.ListChart)
	klog.V(6).Infof("注册 Helm 插件管理路由(mgm)")
}
