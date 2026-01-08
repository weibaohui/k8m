package gatewayapi

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/gatewayapi/route"
)

var Metadata = plugins.Module{
	Meta: plugins.Meta{
		Name:        modules.PluginNameGatewayAPI,
		Title:       "Gateway API",
		Version:     "1.0.0",
		Description: "Kubernetes Gateway API 管理",
	},
	Tables: []string{},
	Crons:  []string{},
	Menus: []plugins.Menu{
		{
			Key:   "plugin_gatewayapi_index",
			Title: "Gateway API",
			Icon:  "fa-solid fa-network-wired",
			Order: 10,
			Children: []plugins.Menu{
				{
					Key:         "plugin_gatewayapi_gateway",
					Title:       "Gateway",
					Icon:        "fa-solid fa-server",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/gatewayapi/gateway")`,
					Order:       100,
				},
				{
					Key:         "plugin_gatewayapi_gateway_class",
					Title:       "GatewayClass",
					Icon:        "fa-solid fa-layer-group",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/gatewayapi/gateway_class")`,
					Order:       101,
				},
				{
					Key:         "plugin_gatewayapi_http_route",
					Title:       "HTTPRoute",
					Icon:        "fa-solid fa-route",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/gatewayapi/http_route")`,
					Order:       102,
				},
				{
					Key:         "plugin_gatewayapi_grpc_route",
					Title:       "GRPCRoute",
					Icon:        "fa-solid fa-route",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/gatewayapi/grpc_route")`,
					Order:       103,
				},
				{
					Key:         "plugin_gatewayapi_tls_route",
					Title:       "TLSRoute",
					Icon:        "fa-solid fa-route",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/gatewayapi/tls_route")`,
					Order:       104,
				},
				{
					Key:         "plugin_gatewayapi_tcp_route",
					Title:       "TCPRoute",
					Icon:        "fa-solid fa-route",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/gatewayapi/tcp_route")`,
					Order:       105,
				},
				{
					Key:         "plugin_gatewayapi_udp_route",
					Title:       "UDPRoute",
					Icon:        "fa-solid fa-route",
					EventType:   "custom",
					CustomEvent: `() => loadJsonPage("/plugins/gatewayapi/udp_route")`,
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
