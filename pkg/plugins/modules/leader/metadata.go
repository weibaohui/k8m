package leader

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
)

// Metadata Leader选举插件的元信息与能力声明
var Metadata = plugins.Module{
	Meta: plugins.Meta{
		Name:        modules.PluginNameLeader,
		Title:       "主备高可用插件",
		Version:     "1.0.0",
		Description: "提供主备高可用能力：多实例部署时仅一个为主节点，通过 Kubernetes 原生机制完成选主。使用前请务必启用 /health/ready 就绪探针。",
	},
	Tables:    []string{},
	Lifecycle: &LeaderLifecycle{},
}
