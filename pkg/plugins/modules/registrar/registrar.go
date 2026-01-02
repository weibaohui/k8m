package registrar

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules/demo"
	"github.com/weibaohui/k8m/pkg/plugins/modules/eventhandler"
	"github.com/weibaohui/k8m/pkg/plugins/modules/gllog"
	"github.com/weibaohui/k8m/pkg/plugins/modules/helm"
	"github.com/weibaohui/k8m/pkg/plugins/modules/inspection"
	"github.com/weibaohui/k8m/pkg/plugins/modules/leader"
	"github.com/weibaohui/k8m/pkg/plugins/modules/mcp"
	"github.com/weibaohui/k8m/pkg/plugins/modules/swagger"
	"github.com/weibaohui/k8m/pkg/plugins/modules/webhook"
	"k8s.io/klog/v2"
)

// init 插件集中注册器
// 在系统启动时设置plugins的集中注册函数，统一注册各插件
func init() {
	plugins.SetRegistrar(func(m *plugins.Manager) {
		if err := m.Register(demo.Metadata); err != nil {
			klog.V(6).Infof("注册demo插件失败: %v", err)
		} else {
			klog.V(6).Infof("注册demo插件成功")
		}
		if err := m.Register(leader.Metadata); err != nil {
			klog.V(6).Infof("注册leader插件失败: %v", err)
		} else {
			klog.V(6).Infof("注册leader插件成功")
		}
		if err := m.Register(webhook.Metadata); err != nil {
			klog.V(6).Infof("注册webhook插件失败: %v", err)
		} else {
			klog.V(6).Infof("注册webhook插件成功")
		}
		if err := m.Register(eventhandler.Metadata); err != nil {
			klog.V(6).Infof("注册eventhandler插件失败: %v", err)
		} else {
			klog.V(6).Infof("注册eventhandler插件成功")
		}
		if err := m.Register(inspection.Metadata); err != nil {
			klog.V(6).Infof("注册inspection插件失败: %v", err)
		} else {
			klog.V(6).Infof("注册inspection插件成功")
		}
		if err := m.Register(helm.Metadata); err != nil {
			klog.V(6).Infof("注册helm插件失败: %v", err)
		} else {
			klog.V(6).Infof("注册helm插件成功")
		}
		if err := m.Register(gllog.Metadata); err != nil {
			klog.V(6).Infof("注册gllog插件失败: %v", err)
		} else {
			klog.V(6).Infof("注册gllog插件成功")
		}
		if err := m.Register(swagger.Metadata); err != nil {
			klog.V(6).Infof("注册swagger插件失败: %v", err)
		} else {
			klog.V(6).Infof("注册swagger插件成功")
		}
		if err := m.Register(mcp.Metadata); err != nil {
			klog.V(6).Infof("注册mcp插件失败: %v", err)
		} else {
			klog.V(6).Infof("注册mcp插件成功")
		}
	})
}
