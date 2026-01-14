# PersistentVolume 合规性检查

## 介绍

检测 PV 是否为 Released/Failed 状态，及容量小于 1Gi。

## 信息

- ScriptCode: Builtin_PV_028
- Kind: PersistentVolume
- Group: core
- Version: v1
- TimeoutSeconds: 45

## 代码

```lua

			local pvs, err = kubectl:GVK("", "v1", "PersistentVolume"):AllNamespace(""):List()
			if err then print("获取 PersistentVolume 失败: " .. tostring(err)) return end
			for _, pv in ipairs(pvs) do
				if pv.status and pv.status.phase == "Released" then
					check_event("失败", "PersistentVolume '" .. pv.metadata.name .. "' 处于 Released 状态，应及时清理", {name=pv.metadata.name})
				end
				if pv.status and pv.status.phase == "Failed" then
					check_event("失败", "PersistentVolume '" .. pv.metadata.name .. "' 处于 Failed 状态", {name=pv.metadata.name})
				end
				if pv.spec and pv.spec.capacity and pv.spec.capacity.storage then
					local function parseGi(val)
						local n = tonumber(val:match("%d+"))
						if val:find("Gi") then return n end
						if val:find("Mi") then return n and n/1024 or 0 end
						return 0
					end
					if parseGi(pv.spec.capacity.storage) < 1 then
						check_event("失败", "PersistentVolume '" .. pv.metadata.name .. "' 容量过小 (" .. pv.spec.capacity.storage .. ")", {name=pv.metadata.name, capacity=pv.spec.capacity.storage})
					end
				end
			end
			print("PersistentVolume 合规性检查完成")
		
```
