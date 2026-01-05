package leader

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
)

// Metadata Leader选举插件的元信息与能力声明
var Metadata = plugins.Module{
	Meta: plugins.Meta{
		Name:        modules.PluginNameLeader,
		Title:       "多实例选举插件",
		Version:     "1.0.0",
		Description: "提供多实例自动选举能力：通过 Kubernetes 原生机制完成选主。使用前请务必启用 /health/ready 就绪探针。启用后访问流量会集中到主实例。",
	},
	Tables:    []string{},
	Lifecycle: &LeaderLifecycle{},
}
