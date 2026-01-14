# Service Selector 检查

## 介绍

检查每个 Service 的 selector 是否有对应 Pod

## 信息

- ScriptCode: Builtin_Service_001
- Kind: Service
- Group: 
- Version: v1
- TimeoutSeconds: 30

## 代码

```lua

		    -- 获取Selector 定义文档
			local doc, err = kubectl:GVK("", "v1", "Service"):Cache(10):Doc("spec.selector")
			if err then
				print( "获取 Service Doc 失败".. tostring(err))
				return
			end
			-- 检查每个 Service 的 selector 是否有对应 Pod，Pod 查询限定在 Service 所在的 namespace
			local svcs, err = kubectl:GVK("", "v1", "Service"):AllNamespace(""):List()
			if not err and svcs then
				for _, svc in ipairs(svcs) do
					if svc.spec and svc.spec.selector then
						local selector = svc.spec.selector
						local labelSelector = ""
						for k, v in pairs(selector) do
							if labelSelector ~= "" then
								labelSelector = labelSelector .. ","
							end
							labelSelector = labelSelector .. k .. "=" .. v
						end
						-- 这里使用 Namespace(svc.metadata.namespace) 保证只查找与 Service 相同命名空间下的 Pod
						local pods, err = kubectl:GVK("", "v1", "Pod"):Namespace(svc.metadata.namespace):Cache(10):WithLabelSelector(labelSelector):List()
						local count = 0
						if not err and pods then
							for _, _ in pairs(pods) do count = count + 1 end
						end
						if count == 0 then
							check_event("失败", "Service " .. svc.metadata.name .. " selector " .. labelSelector .. " 应该至少一个pod, 但是现在没有。" .. "spec.selector定义" .. doc, {name=svc.metadata.name, selector=labelSelector, namespace=svc.metadata.namespace})
						end
					end
				end
			else
				print("Service 列表获取失败: " .. tostring(err))
			end
			print("Service Selector 检查完成")
		
```
