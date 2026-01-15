# PersistentVolumeClaim 合规性检查

## 介绍

检测 PVC Pending/Lost 状态、容量小于 1Gi、无 StorageClass。

## 信息

- ScriptCode: Builtin_PVC_029
- Kind: PersistentVolumeClaim
- Group: core
- Version: v1
- TimeoutSeconds: 0

## 代码

```lua

			local pvcs, err = kubectl:GVK("", "v1", "PersistentVolumeClaim"):AllNamespace(""):List()
			if err then print("获取 PVC 失败: " .. tostring(err)) return end
			for _, pvc in ipairs(pvcs) do
				if pvc.status and pvc.status.phase == "Pending" then
					check_event("失败", "PersistentVolumeClaim '" .. pvc.metadata.name .. "' 处于 Pending 状态", {namespace=pvc.metadata.namespace, name=pvc.metadata.name})
				elseif pvc.status and pvc.status.phase == "Lost" then
					check_event("失败", "PersistentVolumeClaim '" .. pvc.metadata.name .. "' 处于 Lost 状态", {namespace=pvc.metadata.namespace, name=pvc.metadata.name})
				else
					if pvc.spec and pvc.spec.resources and pvc.spec.resources.requests and pvc.spec.resources.requests.storage then
						local function parseGi(val)
							local n = tonumber(val:match("%d+"))
							if val:find("Gi") then return n end
							if val:find("Mi") then return n and n/1024 or 0 end
							return 0
						end
						if parseGi(pvc.spec.resources.requests.storage) < 1 then
							check_event("失败", "PersistentVolumeClaim '" .. pvc.metadata.name .. "' 容量过小 (" .. pvc.spec.resources.requests.storage .. ")", {namespace=pvc.metadata.namespace, name=pvc.metadata.name, capacity=pvc.spec.resources.requests.storage})
						end
					end
					if (not pvc.spec or not pvc.spec.storageClassName) and (not pvc.spec or not pvc.spec.volumeName or pvc.spec.volumeName == "") then
						check_event("失败", "PersistentVolumeClaim '" .. pvc.metadata.name .. "' 未指定 StorageClass", {namespace=pvc.metadata.namespace, name=pvc.metadata.name})
					end
				end
			end
			print("PersistentVolumeClaim 合规性检查完成")
		
```
