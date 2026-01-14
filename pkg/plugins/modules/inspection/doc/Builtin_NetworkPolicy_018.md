# NetworkPolicy 合规性检查

## 介绍

检查 NetworkPolicy 是否允许所有 Pod，或未作用于任何 Pod。

## 信息

- ScriptCode: Builtin_NetworkPolicy_018
- Kind: NetworkPolicy
- Group: networking
- Version: v1
- TimeoutSeconds: 60

## 代码

```lua

			local nps, err = kubectl:GVK("networking.k8s.io", "v1", "NetworkPolicy"):AllNamespace(""):List()
			if err then print("获取 NetworkPolicy 失败: " .. tostring(err)) return end
			for _, np in ipairs(nps) do
				if np.spec and np.spec.podSelector and (not np.spec.podSelector.matchLabels or next(np.spec.podSelector.matchLabels) == nil) then
					check_event("失败", "NetworkPolicy '" .. np.metadata.name .. "' 允许所有 Pod", {namespace=np.metadata.namespace, name=np.metadata.name})
				else
					local selector = ""
					if np.spec and np.spec.podSelector and np.spec.podSelector.matchLabels then
						for k, v in pairs(np.spec.podSelector.matchLabels) do
							if selector ~= "" then selector = selector .. "," end
							selector = selector .. k .. "=" .. v
						end
					end
					if selector ~= "" then
						local pods, err = kubectl:GVK("", "v1", "Pod"):Namespace(np.metadata.namespace):WithLabelSelector(selector):List()
						if not err and pods and #pods.items == 0 then
							check_event("失败", "NetworkPolicy '" .. np.metadata.name .. "' 未作用于任何 Pod", {namespace=np.metadata.namespace, name=np.metadata.name})
						end
					end
				end
			end
			print("NetworkPolicy 合规性检查完成")
		
```
