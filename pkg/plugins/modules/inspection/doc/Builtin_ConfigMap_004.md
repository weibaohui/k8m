# ConfigMap 超大检测

## 介绍

检测所有超过 1MB 的 ConfigMap

## 信息

- ScriptCode: Builtin_ConfigMap_004
- Kind: ConfigMap
- Group: 
- Version: v1
- TimeoutSeconds: 45

## 代码

```lua

			local configmaps, err = kubectl:GVK("", "v1", "ConfigMap"):AllNamespace(""):List()
			if err then
				print( "获取 ConfigMap 失败".. tostring(err))
				return
			end
			for _, cm in ipairs(configmaps) do
				local cmName = cm.metadata.name
				local cmNamespace = cm.metadata.namespace
				local totalSize = 0
				if cm.data then
					for k, v in pairs(cm.data) do
						if type(v) == "string" then
							totalSize = totalSize + string.len(v)
						end
					end
				end
				if cm.binaryData then
					for k, v in pairs(cm.binaryData) do
						if type(v) == "string" then
							totalSize = totalSize + string.len(v)
						end
					end
				end
				local maxSize = 1024 * 1024
				if totalSize > maxSize then
					local sizeMB = string.format("%.2f", totalSize / (1024 * 1024))
					check_event("失败", "[超大] ConfigMap " .. cmNamespace .. "/" .. cmName .. " 大小为 " .. sizeMB .. "MB，超过 1MB 限制", {namespace=cmNamespace, name=cmName, size=sizeMB})
				end
			end
			print("ConfigMap 超大检测完成")
		
```
