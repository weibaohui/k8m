# StorageClass 合规性检查

## 介绍

检测 StorageClass 是否使用了已废弃的 provisioner，及是否存在多个默认 StorageClass。

## 信息

- ScriptCode: Builtin_StorageClass_027
- Kind: StorageClass
- Group: storage.k8s.io
- Version: v1
- TimeoutSeconds: 30

## 代码

```lua

			local scs, err = kubectl:GVK("storage.k8s.io", "v1", "StorageClass"):AllNamespace(""):List()
			if err then print("获取 StorageClass 失败: " .. tostring(err)) return end
			local defaultCount = 0
			for _, sc in ipairs(scs) do
				if sc.provisioner == "kubernetes.io/no-provisioner" then
					check_event("失败", "StorageClass '" .. sc.metadata.name .. "' 使用了已废弃的 provisioner 'kubernetes.io/no-provisioner'", {name=sc.metadata.name})
				end
				if sc.metadata.annotations and sc.metadata.annotations["storageclass.kubernetes.io/is-default-class"] == "true" then
					defaultCount = defaultCount + 1
				end
			end
			if defaultCount > 1 then
				check_event("失败", "存在多个默认 StorageClass (" .. tostring(defaultCount) .. ")，可能导致混淆", {})
			end
			print("StorageClass 合规性检查完成")
		
```
