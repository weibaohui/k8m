package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/yaml_editor/controller"
	"github.com/weibaohui/k8m/pkg/response"
	"k8s.io/klog/v2"
)

func RegisterClusterRoutes(arg chi.Router) {
	prefix := "/plugins/" + modules.PluginNameYamlEditor
	ctrl := &controller.Controller{}
	arg.Post(prefix+"/yaml/apply", response.Adapter(ctrl.Apply))
	arg.Post(prefix+"/yaml/upload", response.Adapter(ctrl.UploadFile))
	arg.Post(prefix+"/yaml/delete", response.Adapter(ctrl.Delete))
	klog.V(6).Infof("注册 YAML 编辑器插件集群路由")
}

func RegisterManagementRoutes(arg chi.Router) {
	prefix := "/plugins/" + modules.PluginNameYamlEditor
	ctrl := &controller.Controller{}
	arg.Get(prefix+"/template/kind/list", response.Adapter(ctrl.ListKind))
	arg.Get(prefix+"/template/list", response.Adapter(ctrl.List))
	arg.Post(prefix+"/template/save", response.Adapter(ctrl.Save))
	arg.Post(prefix+"/template/delete/{ids}", response.Adapter(ctrl.DeleteTemplate))
	klog.V(6).Infof("注册 YAML 编辑器插件管理路由")
}
