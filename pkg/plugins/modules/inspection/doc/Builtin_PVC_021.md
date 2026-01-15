# PVC 合规性检查

## 介绍

检查 PVC Pending 状态下的 ProvisioningFailed 事件。

## 信息

- ScriptCode: Builtin_PVC_021
- Kind: PersistentVolumeClaim
- Group: 
- Version: v1
- TimeoutSeconds: 60

## 代码

```lua

			local pvcs, err = kubectl:GVK("", "v1", "PersistentVolumeClaim"):AllNamespace(""):List()
			if err then print("获取 PVC 失败: " .. tostring(err)) return end
			for _, pvc in ipairs(pvcs) do
				if pvc.status and pvc.status.phase == "Pending" then
					local events, err = kubectl:GVK("", "v1", "Event"):Namespace(pvc.metadata.namespace):WithFieldSelector("involvedObject.name=" .. pvc.metadata.name):List()
					if not err and events and events.items then
						for _, evt in ipairs(events.items) do
							if evt.reason == "ProvisioningFailed" and evt.message and evt.message ~= "" then
								check_event("失败", evt.message, {namespace=pvc.metadata.namespace, name=pvc.metadata.name})
							end
						end
					end
				end
			end
			print("PVC 合规性检查完成")
		
```
