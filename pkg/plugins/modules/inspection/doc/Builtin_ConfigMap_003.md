# ConfigMap 空数据检测

## 介绍

检测所有 data 和 binaryData 字段都为空的 ConfigMap

## 信息

- ScriptCode: Builtin_ConfigMap_003
- Kind: ConfigMap
- Group: 
- Version: v1
- TimeoutSeconds: 30

## 代码

```lua

			local configmaps, err = kubectl:GVK("", "v1", "ConfigMap"):AllNamespace(""):List()
			if err then
				print("获取 ConfigMap 失败".. tostring(err))
				return
			end
			for _, cm in ipairs(configmaps) do
				local cmName = cm.metadata.name
				local cmNamespace = cm.metadata.namespace
				local isEmpty = true
				if cm.data then
					for k, v in pairs(cm.data) do
						isEmpty = false
						break
					end
				end
				if isEmpty and cm.binaryData then
					for k, v in pairs(cm.binaryData) do
						isEmpty = false
						break
					end
				end
				if isEmpty then
					check_event("失败", "[空数据] ConfigMap " .. cmNamespace .. "/" .. cmName .. " 的 data 和 binaryData 字段都为空", {namespace=cmNamespace, name=cmName})
				end
			end
			print("ConfigMap 空数据检测完成")
		
```
