package k8swatch

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
)

var Metadata = plugins.Module{
	Meta: plugins.Meta{
		Name:        modules.PluginNameK8sWatch,
		Title:       "K8s资源监听插件",
		Version:     "1.0.0",
		Description: "监听Kubernetes资源变更，包括Pod、Node、PVC、PV、Ingress等",
	},
	Tables:       []string{},
	Dependencies: []string{},
	RunAfter: []string{
		modules.PluginNameLeader,
	},
	Lifecycle: &K8sWatchLifecycle{},
}
