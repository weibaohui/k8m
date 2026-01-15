# Security ServiceAccount 默认账户使用检测

## 介绍

检测 default ServiceAccount 是否被 Pod 使用。

## 信息

- ScriptCode: Builtin_Security_SA_023
- Kind: ServiceAccount
- Group: core
- Version: v1
- TimeoutSeconds: 60

## 代码

```lua

			local sas, err = kubectl:GVK("", "v1", "ServiceAccount"):AllNamespace(""):List()
			if err then print("获取 ServiceAccount 失败: " .. tostring(err)) return end
			for _, sa in ipairs(sas) do
				if sa.metadata and sa.metadata.name == "default" then
					local pods, err = kubectl:GVK("", "v1", "Pod"):Namespace(sa.metadata.namespace):List()
					if not err and pods then
						local defaultSAUsers = {}
						for _, pod in ipairs(pods) do
							if pod.spec and pod.spec.serviceAccountName == "default" then
								table.insert(defaultSAUsers, pod.metadata.name)
							end
						end
						if #defaultSAUsers > 0 then
							check_event("失败", "Default service account 被以下 Pod 使用: " .. table.concat(defaultSAUsers, ", "), {namespace=sa.metadata.namespace, name=sa.metadata.name})
						end
					end
				end
			end
			print("Security ServiceAccount 检查完成")
		
```
