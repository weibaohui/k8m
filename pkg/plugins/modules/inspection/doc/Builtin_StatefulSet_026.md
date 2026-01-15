# StatefulSet 合规性检查

## 介绍

检测 StatefulSet 关联的 Service、StorageClass 是否存在及 Pod 状态。

## 信息

- ScriptCode: Builtin_StatefulSet_026
- Kind: StatefulSet
- Group: apps
- Version: v1
- TimeoutSeconds: 120

## 代码

```lua

			local stss, err = kubectl:GVK("apps", "v1", "StatefulSet"):AllNamespace(""):List()
			if err then print("获取 StatefulSet 失败: " .. tostring(err)) return end
			for _, sts in ipairs(stss) do
				if sts.spec and sts.spec.serviceName then
					local svc, err = kubectl:GVK("", "v1", "Service"):Namespace(sts.metadata.namespace):Name(sts.spec.serviceName):Get()
					if err or not svc then
						check_event("失败", "StatefulSet 使用的 Service '" .. sts.metadata.namespace .. "/" .. sts.spec.serviceName .. "' 不存在", {namespace=sts.metadata.namespace, name=sts.metadata.name, service=sts.spec.serviceName})
					end
				end
				if sts.spec and sts.spec.volumeClaimTemplates then
					for _, vct in ipairs(sts.spec.volumeClaimTemplates) do
						if vct.spec and vct.spec.storageClassName then
							local sc, err = kubectl:GVK("storage.k8s.io", "v1", "StorageClass"):Name(vct.spec.storageClassName):Get()
							if err or not sc then
								check_event("失败", "StatefulSet 使用的 StorageClass '" .. vct.spec.storageClassName .. "' 不存在", {namespace=sts.metadata.namespace, name=sts.metadata.name, storageClass=vct.spec.storageClassName})
							end
						end
					end
				end
				if sts.spec and sts.spec.replicas and sts.status and sts.status.availableReplicas and sts.spec.replicas ~= sts.status.availableReplicas then
					for i = 0, sts.spec.replicas - 1 do
						local podName = sts.metadata.name .. "-" .. tostring(i)
						local pod, err = kubectl:GVK("", "v1", "Pod"):Namespace(sts.metadata.namespace):Name(podName):Get()
						if err or not pod then
							if i == 0 then
								local events, err = kubectl:GVK("", "v1", "Event"):Namespace(sts.metadata.namespace):WithFieldSelector("involvedObject.name=" .. sts.metadata.name):List()
								if not err and events and events.items then
									for _, evt in ipairs(events.items) do
										if evt.type ~= "Normal" and evt.message and evt.message ~= "" then
											check_event("失败", evt.message, {namespace=sts.metadata.namespace, name=sts.metadata.name})
										end
									end
								end
							end
							break
						end
						if pod.status and pod.status.phase ~= "Running" then
							check_event("失败", "StatefulSet 的 Pod '" .. pod.metadata.name .. "' 不在 Running 状态", {namespace=sts.metadata.namespace, name=sts.metadata.name, pod=pod.metadata.name, phase=pod.status.phase})
							break
						end
					end
				end
			end
			print("StatefulSet 合规性检查完成")
		
```
