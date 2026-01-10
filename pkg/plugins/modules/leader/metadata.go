package leader

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
)

// Metadata Leader选举插件的元信息与能力声明
var Metadata = plugins.Module{
	Meta: plugins.Meta{
		Name:        modules.PluginNameLeader,
		Title:       "多实例主备选举插件",
		Version:     "1.0.0",
		Description: "提供多实例自动选举能力：通过 Kubernetes 原生机制完成选主。通过LabelSelector筛选具有k8m.io/role: leader的Pod为访问流量承载Pod。其他Pod不承载访问流量。",
	},
	Tables:    []string{},
	Lifecycle: &LeaderLifecycle{},
}
