package istio

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/istio/route"
)

var Metadata = plugins.Module{
	Meta: plugins.Meta{
		Name:        modules.PluginNameIstio,
		Title:       "Istio管理插件",
		Version:     "1.0.0",
		Description: "Kubernetes Istio 服务网格管理",
	},
	Tables: []string{},
	Crons:  []string{},
	Menus: []plugins.Menu{
		{
			Key:   "plugin_istio_index",
			Title: "Istio",
			Icon:  "fa-solid fa-cube",
			Order: 9,
			Children: []plugins.Menu{
				{
					Key:         "plugin_istio_virtual_service",
					Title:       "虚拟服务",
					Icon:        "fa-solid fa-route",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/istio/VirtualService")`,
					Order:       100,
				},
				{
					Key:         "plugin_istio_destination_rule",
					Title:       "目标规则",
					Icon:        "fa-solid fa-location-dot",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/istio/DestinationRule")`,
					Order:       101,
				},
				{
					Key:         "plugin_istio_envoy_filter",
					Title:       "Envoy过滤器",
					Icon:        "fa-solid fa-filter",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/istio/EnvoyFilter")`,
					Order:       102,
				},
				{
					Key:         "plugin_istio_gateway",
					Title:       "网关",
					Icon:        "fa-solid fa-network-wired",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/istio/Gateway")`,
					Order:       103,
				},
				{
					Key:         "plugin_istio_peer_authentication",
					Title:       "对等认证",
					Icon:        "fa-solid fa-user-shield",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/istio/PeerAuthentication")`,
					Order:       104,
				},
				{
					Key:         "plugin_istio_proxy_config",
					Title:       "代理配置",
					Icon:        "fa-solid fa-gears",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/istio/ProxyConfig")`,
					Order:       105,
				},
				{
					Key:         "plugin_istio_request_authentication",
					Title:       "请求认证",
					Icon:        "fa-solid fa-key",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/istio/RequestAuthentication")`,
					Order:       106,
				},
				{
					Key:         "plugin_istio_service_entry",
					Title:       "服务入口",
					Icon:        "fa-solid fa-door-open",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/istio/ServiceEntry")`,
					Order:       107,
				},
				{
					Key:         "plugin_istio_sidecar",
					Title:       "边车",
					Icon:        "fa-solid fa-car-side",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/istio/Sidecar")`,
					Order:       108,
				},
				{
					Key:         "plugin_istio_telemetry",
					Title:       "遥测",
					Icon:        "fa-solid fa-chart-line",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/istio/Telemetry")`,
					Order:       109,
				},
				{
					Key:         "plugin_istio_authorization_policy",
					Title:       "授权策略",
					Icon:        "fa-solid fa-user-lock",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/istio/AuthorizationPolicy")`,
					Order:       110,
				},
				{
					Key:         "plugin_istio_wasm_plugin",
					Title:       "Wasm插件",
					Icon:        "fa-solid fa-puzzle-piece",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/istio/WasmPlugin")`,
					Order:       111,
				},
				{
					Key:         "plugin_istio_workload_entry",
					Title:       "工作负载条目",
					Icon:        "fa-solid fa-server",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/istio/WorkloadEntry")`,
					Order:       112,
				},
				{
					Key:         "plugin_istio_workload_group",
					Title:       "工作负载组",
					Icon:        "fa-solid fa-people-group",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/istio/WorkloadGroup")`,
					Order:       113,
				},
			},
		},
	},
	Dependencies:  []string{},
	RunAfter:      []string{},
	Lifecycle:     &IstioLifecycle{},
	ClusterRouter: route.RegisterClusterRoutes,
}
