package openkruise

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/openkruise/route"
)

var Metadata = plugins.Module{
	Meta: plugins.Meta{
		Name:        modules.PluginNameOpenKruise,
		Title:       "OpenKruise管理插件",
		Version:     "1.0.0",
		Description: "Kubernetes OpenKruise 高级工作负载管理",
	},
	Tables: []string{},
	Crons:  []string{},
	Menus: []plugins.Menu{
		{
			Key:   "OpenKruise-workload",
			Title: "OpenKruise",
			Icon:  "fa-solid fa-cube",
			Order: 8,
			Children: []plugins.Menu{
				{
					Key:         "advanced-cloneset",
					Title:       "克隆集",
					Icon:        "fa-solid fa-clone",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/openkruise/cloneset")`,
					Order:       1,
				},
				{
					Key:         "advanced-statefulset",
					Title:       "高级有状态集",
					Icon:        "fa-solid fa-layer-group",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/openkruise/statefulset")`,
					Order:       2,
				},
				{
					Key:         "advanced-daemonSet",
					Title:       "高级守护进程集",
					Icon:        "fa-solid fa-shield-halved",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/openkruise/daemonset")`,
					Order:       3,
				},
				{
					Key:         "advanced-cronJob",
					Title:       "高级定时任务",
					Icon:        "fa-solid fa-clock",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/openkruise/cronjob")`,
					Order:       4,
				},
				{
					Key:         "broadcast-job",
					Title:       "广播作业任务",
					Icon:        "fa-solid fa-broadcast-tower",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/openkruise/broadcastjob")`,
					Order:       5,
				},
				{
					Key:         "sidecarset",
					Title:       "边车集",
					Icon:        "fa-solid fa-car-side",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/openkruise/sidecarset")`,
					Order:       6,
				},
				{
					Key:         "workload-spread",
					Title:       "工作负载分布",
					Icon:        "fa-solid fa-arrows-split-up-and-left",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/openkruise/workloadspread")`,
					Order:       7,
				},
				{
					Key:         "united-deployment",
					Title:       "联合部署",
					Icon:        "fa-solid fa-object-group",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/openkruise/uniteddeployment")`,
					Order:       8,
				},
				{
					Key:         "container_recreate_request",
					Title:       "容器重建请求",
					Icon:        "fa-solid fa-recycle",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/openkruise/container_recreate_request")`,
					Order:       9,
				},
				{
					Key:         "imagepulljob",
					Title:       "镜像拉取作业",
					Icon:        "fa-solid fa-cloud-arrow-down",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/openkruise/imagepulljob")`,
					Order:       10,
				},
				{
					Key:         "persistentpodstate",
					Title:       "持久化状态",
					Icon:        "fa-solid fa-database",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/openkruise/persistentpodstate")`,
					Order:       11,
				},
				{
					Key:         "podprobemarker",
					Title:       "Pod探测标记",
					Icon:        "fa-solid fa-magnifying-glass",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/openkruise/podprobemarker")`,
					Order:       12,
				},
				{
					Key:         "PodUnavailableBudget",
					Title:       "Pod不可用预算",
					Icon:        "fa-solid fa-circle-exclamation",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/openkruise/PodUnavailableBudget")`,
					Order:       13,
				},
				{
					Key:         "ResourceDistribution",
					Title:       "资源分发",
					Icon:        "fa-solid fa-share-nodes",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/openkruise/ResourceDistribution")`,
					Order:       14,
				},
			},
		},
	},
	Dependencies:  []string{},
	RunAfter:      []string{},
	Lifecycle:     &OpenKruiseLifecycle{},
	ClusterRouter: route.RegisterClusterRoutes,
}
