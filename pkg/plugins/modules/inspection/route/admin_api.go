package route

import (
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/inspection/controller"
	"github.com/weibaohui/k8m/pkg/response"
	"k8s.io/klog/v2"
)

// RegisterPluginAdminRoutes 注册集群巡检插件的管理员路由 - Gin到Chi迁移
// 使用插件内部的 controller 包，完全自包含
func RegisterPluginAdminRoutes(arg chi.Router) {
	prefix := "/plugins/" + modules.PluginNameInspection

	ctrl := &controller.AdminScheduleController{}
	arg.Get(prefix+"/schedule/list", response.Adapter(ctrl.List))
	arg.Get(prefix+"/schedule/record/id/{id}/event/list", response.Adapter(ctrl.EventList))
	arg.Post(prefix+"/schedule/record/id/{id}/summary", response.Adapter(ctrl.SummaryByRecordID))
	arg.Get(prefix+"/schedule/record/id/{id}/output/list", response.Adapter(ctrl.OutputList))
	arg.Post(prefix+"/schedule/save", response.Adapter(ctrl.Save))
	arg.Post(prefix+"/schedule/delete/{ids}", response.Adapter(ctrl.Delete))
	arg.Post(prefix+"/schedule/save/id/{id}/status/{enabled}", response.Adapter(ctrl.QuickSave))
	arg.Post(prefix+"/schedule/start/id/{id}", response.Adapter(ctrl.Start))
	arg.Post(prefix+"/schedule/id/{id}/update_script_code", response.Adapter(ctrl.UpdateScriptCode))
	arg.Post(prefix+"/schedule/id/{id}/summary", response.Adapter(ctrl.SummaryBySchedule))
	arg.Post(prefix+"/schedule/id/{id}/summary/cluster/{cluster}/start_time/{start_time}/end_time/{end_time}", response.Adapter(ctrl.SummaryBySchedule))
	arg.Get(prefix+"/event/status/option_list", response.Adapter(ctrl.EventStatusOptionList))

	rc := &controller.AdminRecordController{}
	arg.Get(prefix+"/schedule/id/{id}/record/list", response.Adapter(rc.RecordList))
	arg.Get(prefix+"/record/list", response.Adapter(rc.RecordList))
	arg.Post(prefix+"/schedule/record/id/{id}/push", response.Adapter(rc.Push))

	sc := &controller.AdminLuaScriptController{}
	arg.Get(prefix+"/script/list", response.Adapter(sc.LuaScriptList))
	arg.Post(prefix+"/script/delete/{ids}", response.Adapter(sc.LuaScriptDelete))
	arg.Post(prefix+"/script/save", response.Adapter(sc.LuaScriptSave))
	arg.Post(prefix+"/script/load", response.Adapter(sc.LuaScriptLoad))
	arg.Get(prefix+"/script/option_list", response.Adapter(sc.LuaScriptOptionList))

	klog.V(6).Infof("注册集群巡检插件管理路由(admin)")
}
