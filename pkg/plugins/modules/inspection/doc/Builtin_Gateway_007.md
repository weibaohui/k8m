# Gateway 合规性检查

## 介绍

检查 Gateway 关联的 GatewayClass 是否存在，以及 Gateway 状态是否被接受

## 信息

- ScriptCode: Builtin_Gateway_007
- Kind: Gateway
- Group: gateway.networking.k8s.io
- Version: v1
- TimeoutSeconds: 45

## 代码

```lua

			local gateways, err = kubectl:GVK("gateway.networking.k8s.io", "v1", "Gateway"):AllNamespace(""):List()
			if err then
				print("获取 Gateway 失败: " .. tostring(err))
				return
			end
			for _, gtw in ipairs(gateways) do
				local ns = gtw.metadata and gtw.metadata.namespace or "default"
				local name = gtw.metadata and gtw.metadata.name or ""
				local className = gtw.spec and gtw.spec.gatewayClassName or nil
				local classExists = false
				if className then
					local gtwclass, err = kubectl:GVK("gateway.networking.k8s.io", "v1", "GatewayClass"):Name(className):Get()
					if not err and gtwclass then
						classExists = true
					end
				end
				if not classExists then
					check_event("失败", "Gateway 使用的 GatewayClass " .. tostring(className) .. " 不存在", {namespace=ns, name=name, gatewayClassName=className})
				end
				-- 检查第一个 Condition 状态
				if gtw.status and gtw.status.conditions and #gtw.status.conditions > 0 then
					local cond = gtw.status.conditions[1]
					if cond.status ~= "True" then
						check_event("失败", "Gateway '" .. ns .. "/" .. name .. "' 未被接受, Message: '" .. (cond.message or "") .. "'", {namespace=ns, name=name, message=cond.message})
					end
				end
			end
			print("Gateway 合规性检查完成")
		
```
