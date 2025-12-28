package eventhandler

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/eventhandler/worker"
)

// NotifyPlatformConfigUpdated 中文函数注释：当平台参数（事件转发开关/批次/周期等）更新后，通知插件立即刷新配置并尽快生效。
func NotifyPlatformConfigUpdated() {
	if !plugins.ManagerInstance().IsEnabled(modules.PluginNameEventHandler) {
		return
	}
	if w := worker.Instance(); w != nil {
		w.UpdateConfig()
	}
	SyncEventForwardingFromConfig()
}

