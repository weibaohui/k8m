package registrar

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules/demo"
	"k8s.io/klog/v2"
)

// init 插件集中注册器
// 在系统启动时设置plugins的集中注册函数，统一注册各插件
func init() {
	plugins.SetRegistrar(func(m *plugins.Manager) {
		if err := m.Register(demo.ModuleDef); err != nil {
			klog.V(6).Infof("注册demo插件失败: %v", err)
		} else {
			klog.V(6).Infof("注册demo插件成功")
		}
	})
}

