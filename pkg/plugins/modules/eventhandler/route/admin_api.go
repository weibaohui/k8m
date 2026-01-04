package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/eventhandler/admin"
	"github.com/weibaohui/k8m/pkg/response"
	"k8s.io/klog/v2"
)

// RegisterPluginAdminRoutes 中文函数注释：注册事件转发插件的管理员路由（平台管理员） - Gin到Chi迁移
func RegisterPluginAdminRoutes(arg chi.Router) {
	ctrl := &admin.Controller{}
	prefix := "/plugins/" + modules.PluginNameEventHandler

	arg.Get(prefix+"/setting/get", response.Adapter(ctrl.GetSetting))
	arg.Post(prefix+"/setting/update", response.Adapter(ctrl.UpdateSetting))

	arg.Get(prefix+"/list", response.Adapter(ctrl.List))
	arg.Post(prefix+"/save", response.Adapter(ctrl.Save))
	arg.Post(prefix+"/delete/{ids}", response.Adapter(ctrl.Delete))
	arg.Post(prefix+"/save/id/{id}/status/{enabled}", response.Adapter(ctrl.QuickSave))

	klog.V(6).Infof("注册事件转发插件管理路由(admin)")
}
