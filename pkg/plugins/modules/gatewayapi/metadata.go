package gatewayapi

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/gatewayapi/route"
)

var Metadata = plugins.Module{
	Meta: plugins.Meta{
		Name:        modules.PluginNameGatewayAPI,
		Title:       "Gateway API管理插件",
		Version:     "1.0.0",
		Description: "Kubernetes Gateway API 管理",
	},
	Tables: []string{},
	Crons:  []string{},
	Menus: []plugins.Menu{
		{
			Key:   "plugin_gatewayapi_index",
			Title: "网关API",
			Icon:  "fa-solid fa-door-closed",
			Order: 10,
			Children: []plugins.Menu{
				{
					Key:         "plugin_gatewayapi_gateway_class",
					Title:       "网关类",
					Icon:        "fa-solid fa-door-open",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/gatewayapi/gateway_class")`,
					Order:       100,
				},
				{
					Key:         "plugin_gatewayapi_gateway",
					Title:       "网关",
					Icon:        "fa-solid fa-archway",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/gatewayapi/gateway")`,
					Order:       101,
				},
				{
					Key:         "plugin_gatewayapi_http_route",
					Title:       "HTTP路由",
					Icon:        "fa-solid fa-route",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/gatewayapi/http_route")`,
					Order:       102,
				},
				{
					Key:         "plugin_gatewayapi_grpc_route",
					Title:       "GRPC路由",
					Icon:        "fa-solid fa-code-branch",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/gatewayapi/grpc_route")`,
					Order:       103,
				},
				{
					Key:         "plugin_gatewayapi_tcp_route",
					Title:       "TCP路由",
					Icon:        "fa-solid fa-plug",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/gatewayapi/tcp_route")`,
					Order:       104,
				},
				{
					Key:         "plugin_gatewayapi_udp_route",
					Title:       "UDP路由",
					Icon:        "fa-solid fa-broadcast-tower",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/gatewayapi/udp_route")`,
					Order:       105,
				},
				{
					Key:         "plugin_gatewayapi_tls_route",
					Title:       "TLS路由",
					Icon:        "fa-solid fa-shield-alt",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/gatewayapi/tls_route")`,
					Order:       106,
				},
			},
		},
	},
	Dependencies:  []string{},
	RunAfter:      []string{},
	Lifecycle:     &GatewayAPILifecycle{},
	ClusterRouter: route.RegisterClusterRoutes,
}
