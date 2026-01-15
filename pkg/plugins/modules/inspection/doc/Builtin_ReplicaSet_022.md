# ReplicaSet 合规性检查

## 介绍

检测副本数为0且有 FailedCreate 的 ReplicaFailure。

## 信息

- ScriptCode: Builtin_ReplicaSet_022
- Kind: ReplicaSet
- Group: apps
- Version: v1
- TimeoutSeconds: 45

## 代码

```lua

			local rss, err = kubectl:GVK("apps", "v1", "ReplicaSet"):AllNamespace(""):List()
			if err then print("获取 ReplicaSet 失败: " .. tostring(err)) return end
			for _, rs in ipairs(rss) do
				if rs.status and rs.status.replicas == 0 and rs.status.conditions then
					for _, cond in ipairs(rs.status.conditions) do
						if cond.type == "ReplicaFailure" and cond.reason == "FailedCreate" then
							check_event("失败", cond.message or "ReplicaSet 副本创建失败", {namespace=rs.metadata.namespace, name=rs.metadata.name})
						end
					end
				end
			end
			print("ReplicaSet 合规性检查完成")
		
```
