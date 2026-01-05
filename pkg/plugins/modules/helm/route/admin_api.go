package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/helm/admin"
	"github.com/weibaohui/k8m/pkg/response"
	"k8s.io/klog/v2"
)

// RegisterPluginAdminRoutes 注册 Helm 插件的管理员路由（平台管理员）
// 对应原来的 /admin 路径下的路由
// 从 gin 切换到 chi，使用直接路由方法替代 gin.Group，使用小写方法名
func RegisterPluginAdminRoutes(arg chi.Router) {
	prefix := "/plugins/" + modules.PluginNameHelm
	ctrl := &admin.RepoController{}
	arg.Get(prefix+"/repo/list", response.Adapter(ctrl.List))
	arg.Post(prefix+"/repo/delete/{ids}", response.Adapter(ctrl.Delete))
	arg.Post(prefix+"/repo/update_index", response.Adapter(ctrl.UpdateReposIndex))
	arg.Post(prefix+"/repo/save", response.Adapter(ctrl.Save))

	settingCtrl := &admin.SettingController{}
	arg.Get(prefix+"/setting/get", response.Adapter(settingCtrl.GetSetting))
	arg.Post(prefix+"/setting/update", response.Adapter(settingCtrl.UpdateSetting))

	klog.V(6).Infof("注册 Helm 插件管理路由(admin)")
}

// RegisterPluginAPIRoutes 注册 Helm 插件的 API 路由（K8s API）
// 对应原来的 /k8s/cluster/{cluster} 路径下的路由
// 从 gin 切换到 chi，使用直接路由方法替代 gin.Group，使用小写方法名
func RegisterPluginAPIRoutes(arg chi.Router) {
	prefix := "/plugins/" + modules.PluginNameHelm
	ctrl := &admin.ReleaseController{}

	arg.Get(prefix+"/release/list", response.Adapter(ctrl.ListRelease))
	arg.Get(prefix+"/release/ns/{ns}/name/{name}/history/list", response.Adapter(ctrl.ListReleaseHistory))
	arg.Post(prefix+"/release/{release}/repo/{repo}/chart/{chart}/version/{version}/install", response.Adapter(ctrl.InstallRelease))
	arg.Post(prefix+"/release/ns/{ns}/name/{name}/uninstall", response.Adapter(ctrl.UninstallRelease))
	arg.Get(prefix+"/release/ns/{ns}/name/{name}/revision/{revision}/values", response.Adapter(ctrl.GetReleaseValues))
	arg.Get(prefix+"/release/ns/{ns}/name/{name}/revision/{revision}/notes", response.Adapter(ctrl.GetReleaseNote))
	arg.Get(prefix+"/release/ns/{ns}/name/{name}/revision/{revision}/install_log", response.Adapter(ctrl.GetReleaseInstallLog))
	arg.Post(prefix+"/release/batch/uninstall", response.Adapter(ctrl.BatchUninstallRelease))
	arg.Post(prefix+"/release/upgrade", response.Adapter(ctrl.UpgradeRelease))

	klog.V(6).Infof("注册 Helm 插件 API 路由(api)")
}

// RegisterPluginMgmRoutes 注册 Helm 插件的管理路由（Mgm）
// 对应原来的 /mgm 路径下的路由
// 从 gin 切换到 chi，使用直接路由方法替代 gin.Group，使用小写方法名
func RegisterPluginMgmRoutes(arg chi.Router) {
	prefix := "/plugins/" + modules.PluginNameHelm
	ctrl := &admin.ChartController{}
	arg.Get(prefix+"/repo/{repo}/chart/{chart}/versions", response.Adapter(ctrl.ChartVersionOptionList))
	arg.Get(prefix+"/repo/{repo}/chart/{chart}/version/{version}/values", response.Adapter(ctrl.GetChartValue))
	arg.Get(prefix+"/chart/list", response.Adapter(ctrl.ListChart))
	klog.V(6).Infof("注册 Helm 插件管理路由(mgm)")
}
