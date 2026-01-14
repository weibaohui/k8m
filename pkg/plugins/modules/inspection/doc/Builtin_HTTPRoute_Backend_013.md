# HTTPRoute Backend Service 存在性与端口检查

## 介绍

检查 HTTPRoute 所引用的后端 Service 是否存在，以及端口是否匹配 Service 的端口。

## 信息

- ScriptCode: Builtin_HTTPRoute_Backend_013
- Kind: HTTPRoute
- Group: gateway.networking.k8s.io
- Version: v1
- TimeoutSeconds: 60

## 代码

```lua

			local httproutes, err = kubectl:GVK("gateway.networking.k8s.io", "v1", "HTTPRoute"):AllNamespace(""):List()
			if err then print("获取 HTTPRoute 失败: " .. tostring(err)) return end
			for _, route in ipairs(httproutes) do
				if route.spec and route.spec.rules then
					for _, rule in ipairs(route.spec.rules) do
						if rule.backendRefs then
							for _, backend in ipairs(rule.backendRefs) do
								local svc, err = kubectl:GVK("", "v1", "Service"):Namespace(route.metadata.namespace):Name(backend.name):Get()
								if err or not svc then
									check_event("失败", "HTTPRoute 使用的 Service '" .. route.metadata.namespace .. "/" .. backend.name .. "' 不存在", {namespace=route.metadata.namespace, name=backend.name})
								else
									local portMatch = false
									if svc.spec and svc.spec.ports and backend.port then
										for _, svcPort in ipairs(svc.spec.ports) do
											if svcPort.port == backend.port then portMatch = true end
										end
									end
									if not portMatch then
										check_event("失败", "HTTPRoute 的后端 Service '" .. backend.name .. "' 使用端口 '" .. tostring(backend.port) .. "'，但 Service 未配置该端口", {namespace=route.metadata.namespace, name=backend.name, port=backend.port})
									end
								end
							end
						end
					end
				end
			end
			print("HTTPRoute Backend Service 检查完成")
		
```
