# Pod 日志错误检测

## 介绍

检查某一个 Pod 的最近日志是否包含指定关键字，若包含则认为检测失败。

## 信息

- ScriptCode: Builtin_Pod_Log_Error_031
- Kind: Pod
- Group: 
- Version: v1
- TimeoutSeconds: 0

## 代码

```lua

			-- 示例：根据已知 Deployment 名称与命名空间，按其 selector 获取 Pod 列表并检查日志
			-- 请按需修改以下四个变量
			local deployName = "your-deploy-name"
			local namespace = "default"
			local keyword = "ERROR"  -- 默认关键字为 "ERROR"，可按需改为要检测的关键字
			local tailLines = 200    -- 默认读取最近 200 行日志，按需调整

			-- 获取 Deployment 对象
			local dep, derr = kubectl:GVK("apps", "v1", "Deployment"):Namespace(namespace):Name(deployName):Get()
			if derr ~= nil or not dep then
				print("获取 Deployment 失败: " .. tostring(derr))
				return
			end

			-- 从 Deployment 的 selector.matchLabels 构建 LabelSelector
			local matchLabels = dep.spec and dep.spec.selector and dep.spec.selector.matchLabels or nil
			if not matchLabels then
				print("Deployment 未定义 selector.matchLabels，无法按标签筛选 Pod")
				return
			end
			local labelSelector = ""
			for k, v in pairs(matchLabels) do
				if labelSelector ~= "" then labelSelector = labelSelector .. "," end
				labelSelector = labelSelector .. k .. "=" .. v
			end

			-- 按 Deployment 的 selector 在同一命名空间获取 Pod 列表
			local pods, perr = kubectl:GVK("", "v1", "Pod"):Namespace(namespace):Cache(10):WithLabelSelector(labelSelector):List()
			if perr ~= nil then
				print("获取 Pod 列表失败: " .. tostring(perr))
				return
			end
			if not pods or #pods == 0 then
				print("未找到与 Deployment 匹配的 Pod: " .. namespace .. "/" .. deployName .. ", selector=" .. labelSelector)
				return
			end

			local foundError = false
			for _, pod in ipairs(pods) do
				local ns = pod.metadata.namespace
				local name = pod.metadata.name
				local containerName = nil
				if pod.spec and pod.spec.containers and #pod.spec.containers > 0 then
					containerName = pod.spec.containers[1].name
				end
				local opts = { tailLines = tailLines }
				if containerName ~= nil then
					opts.container = containerName
				end
				local logs, lerr = kubectl:GVK("", "v1", "Pod"):Namespace(ns):Name(name):GetLogs(opts)
				if lerr ~= nil then
					print("获取 Pod 日志失败: " .. tostring(lerr))
				else
					local logStr = (type(logs) == "string") and logs or tostring(logs)
					if logStr and string.find(logStr, keyword) ~= nil then
						foundError = true
						check_event("失败", "Pod 日志包含关键字 '" .. keyword .. "'", {namespace=ns, name=name, container=containerName, keyword=keyword})
					else
						print("Pod " .. ns .. "/" .. name .. " 最近日志未发现 '" .. keyword .. "'")
					end
				end
			end

			if not foundError then
				print("Deployment '" .. namespace .. "/" .. deployName .. "' 关联 Pod 的日志检查完成，未发现 '" .. keyword .. "'")
			end
			print("Pod 日志错误检测完成")
		
```
