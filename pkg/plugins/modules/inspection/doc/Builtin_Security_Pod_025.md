# Security Pod 安全上下文检测

## 介绍

检测 Pod 是否存在特权容器或缺少安全上下文。

## 信息

- ScriptCode: Builtin_Security_Pod_025
- Kind: Pod
- Group: core
- Version: v1
- TimeoutSeconds: 90

## 代码

```lua

			local pods, err = kubectl:GVK("", "v1", "Pod"):AllNamespace(""):List()
			if err then print("获取 Pod 失败: " .. tostring(err)) return end
			for _, pod in ipairs(pods) do
				local hasPrivileged = false
				if pod.spec and pod.spec.containers then
					for _, c in ipairs(pod.spec.containers) do
						if c.securityContext and c.securityContext.privileged == true then
							hasPrivileged = true
							check_event("失败", "容器 " .. c.name .. " 以特权模式运行，存在安全风险", {namespace=pod.metadata.namespace, name=pod.metadata.name, container=c.name})
							break
						end
					end
				end
				if not hasPrivileged and (not pod.spec or not pod.spec.securityContext) then
					check_event("失败", "Pod " .. pod.metadata.name .. " 未定义安全上下文，存在安全风险", {namespace=pod.metadata.namespace, name=pod.metadata.name})
				end
			end
			print("Security Pod 安全上下文检查完成")
		
```
