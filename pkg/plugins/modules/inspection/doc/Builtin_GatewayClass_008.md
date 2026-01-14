# GatewayClass 合规性检查

## 介绍

检查 GatewayClass 的第一个 Condition 状态是否为 True，否则报告未被接受及 message。

## 信息

- ScriptCode: Builtin_GatewayClass_008
- Kind: GatewayClass
- Group: gateway.networking.k8s.io
- Version: v1
- TimeoutSeconds: 45

## 代码

```lua

			local gatewayclasses, err = kubectl:GVK("gateway.networking.k8s.io", "v1", "GatewayClass"):AllNamespace(""):List()
			if err then
				print("获取 GatewayClass 失败: " .. tostring(err))
				return
			end
			for _, gc in ipairs(gatewayclasses) do
				local name = gc.metadata and gc.metadata.name or ""
				if gc.status and gc.status.conditions and #gc.status.conditions > 0 then
					local cond = gc.status.conditions[1]
					if cond.status ~= "True" then
						check_event("失败", "GatewayClass '" .. name .. "' 未被接受, Message: '" .. (cond.message or "") .. "'", {name=name, message=cond.message})
					end
				end
			end
			print("GatewayClass 合规性检查完成")
		
```
