package demo

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"k8s.io/klog/v2"
)

// Register 将Demo插件注册到管理器，并完成安装
func Register(mgr *plugins.Manager) {
	if err := mgr.Register(ModuleDef); err != nil {
		klog.V(6).Infof("注册Demo插件失败: %v", err)
		return
	}
	if err := mgr.Install(ModuleDef.Meta.Name); err != nil {
		klog.V(6).Infof("安装Demo插件失败: %v", err)
		return
	}
}
