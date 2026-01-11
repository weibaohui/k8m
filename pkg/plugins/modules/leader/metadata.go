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
		Description: "提供多实例自动选举能力,在N个实例中选举出一个主实例。使用k8s原生机制完成选主。访问流量应通过LabelSelector筛选带有k8m.io/role: leader标签的Pod做为承载Pod。其他Pod不承载访问流量。使用时请务必注意此点。",
	},
	Tables:    []string{},
	Lifecycle: &LeaderLifecycle{},
}
