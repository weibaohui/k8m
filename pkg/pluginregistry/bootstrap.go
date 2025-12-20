package pluginregistry

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	demoplugin "github.com/weibaohui/k8m/pkg/plugins/modules/demo"
	"k8s.io/klog/v2"
)

// Bootstrap 统一注册所有内置插件，并按默认策略启用
func Bootstrap(mgr *plugins.Manager) {
	// 注册所有模块（集中管理，避免在 main 泄露细节）
	demoplugin.Register(mgr)
	klog.V(6).Infof("完成内置插件注册")

	// 默认策略：启用所有已注册插件（后续可替换为从DB/配置读取启用集）
	mgr.EnableAll()
}

// init 将注册器绑定到插件管理器（仅绑定，不做启停），Manager.Start() 时调用
func init() {
	plugins.SetRegistrar(Bootstrap)
}
