# HTTPRoute Gateway 存在性与命名空间策略检查

## 介绍

检查 HTTPRoute 所引用的 Gateway 是否存在，以及 Gateway 的 AllowedRoutes 策略是否允许该 HTTPRoute。

## 信息

- ScriptCode: Builtin_HTTPRoute_Gateway_014
- Kind: HTTPRoute
- Group: gateway.networking.k8s.io
- Version: v1
- TimeoutSeconds: 75

## 代码

```lua

			local httproutes, err = kubectl:GVK("gateway.networking.k8s.io", "v1", "HTTPRoute"):AllNamespace(""):List()
			if err then print("获取 HTTPRoute 失败: " .. tostring(err)) return end
			for _, route in ipairs(httproutes) do
				if route.spec and route.spec.parentRefs then
					for _, gtwref in ipairs(route.spec.parentRefs) do
						local ns = route.metadata.namespace
						if gtwref.namespace then ns = gtwref.namespace end
						local gtw, err = kubectl:GVK("gateway.networking.k8s.io", "v1", "Gateway"):Namespace(ns):Name(gtwref.name):Get()
						if err or not gtw then
							check_event("失败", "HTTPRoute 使用的 Gateway '" .. ns .. "/" .. gtwref.name .. "' 不存在", {namespace=ns, name=gtwref.name})
						else
							if gtw.spec and gtw.spec.listeners then
								for _, listener in ipairs(gtw.spec.listeners) do
									if listener.allowedRoutes and listener.allowedRoutes.namespaces and listener.allowedRoutes.namespaces.from then
										local allow = listener.allowedRoutes.namespaces.from
										if allow == "Same" and route.metadata.namespace ~= gtw.metadata.namespace then
											check_event("失败", "HTTPRoute '" .. route.metadata.namespace .. "/" .. route.metadata.name .. "' 与 Gateway '" .. gtw.metadata.namespace .. "/" .. gtw.metadata.name .. "' 不在同一命名空间，且 Gateway 只允许同命名空间 HTTPRoute", {route_ns=route.metadata.namespace, route_name=route.metadata.name, gtw_ns=gtw.metadata.namespace, gtw_name=gtw.metadata.name})
										elseif allow == "Selector" and listener.allowedRoutes.namespaces.selector and listener.allowedRoutes.namespaces.selector.matchLabels then
											local match = false
											for k, v in pairs(listener.allowedRoutes.namespaces.selector.matchLabels) do
												if route.metadata.labels and route.metadata.labels[k] == v then match = true end
											end
											if not match then
												check_event("失败", "HTTPRoute '" .. route.metadata.namespace .. "/" .. route.metadata.name .. "' 的标签与 Gateway '" .. gtw.metadata.namespace .. "/" .. gtw.metadata.name .. "' 的 Selector 不匹配", {route_ns=route.metadata.namespace, route_name=route.metadata.name, gtw_ns=gtw.metadata.namespace, gtw_name=gtw.metadata.name})
											end
										end
									end
								end
							end
						end
					end
				end
			end
			print("HTTPRoute Gateway 检查完成")
		
```
