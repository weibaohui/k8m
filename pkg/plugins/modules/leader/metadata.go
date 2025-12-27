package leader

import "github.com/weibaohui/k8m/pkg/plugins"

// Metadata Leader选举插件的元信息与能力声明
var Metadata = plugins.Module{
	Meta: plugins.Meta{
		Name:        "leader",
		Title:       "Leader选举插件",
		Version:     "1.0.0",
		Description: "负责进行Leader选举，并在成为Leader时启动平台的后台任务",
	},
	Lifecycle: &LeaderLifecycle{},
}

