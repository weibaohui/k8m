# ConfigMap 未被使用检测

## 介绍

检测所有未被 Pod 使用的 ConfigMap

## 信息

- ScriptCode: Builtin_ConfigMap_002
- Kind: ConfigMap
- Group: 
- Version: v1
- TimeoutSeconds: 90

## 代码

```lua

			local configmaps, err = kubectl:GVK("", "v1", "ConfigMap"):AllNamespace(""):List()
			if err then
				print("获取 ConfigMap 失败".. tostring(err))
				return
			end
			local pods, err = kubectl:GVK("", "v1", "Pod"):Cache(10):AllNamespace(""):List()
			if err then
				print("获取 Pod 失败".. tostring(err))
				return
			end
			local usedConfigMaps = {}
			for _, pod in ipairs(pods) do
				if pod.spec and pod.spec.volumes then
					for _, volume in ipairs(pod.spec.volumes) do
						if volume.configMap and volume.configMap.name then
							local key = pod.metadata.namespace .. "/" .. volume.configMap.name
							usedConfigMaps[key] = true
						end
					end
				end
				if pod.spec and pod.spec.containers then
					for _, container in ipairs(pod.spec.containers) do
						if container.env then
							for _, env in ipairs(container.env) do
								if env.valueFrom and env.valueFrom.configMapKeyRef and env.valueFrom.configMapKeyRef.name then
									local key = pod.metadata.namespace .. "/" .. env.valueFrom.configMapKeyRef.name
									usedConfigMaps[key] = true
								end
							end
						end
						if container.envFrom then
							for _, envFrom in ipairs(container.envFrom) do
								if envFrom.configMapRef and envFrom.configMapRef.name then
									local key = pod.metadata.namespace .. "/" .. envFrom.configMapRef.name
									usedConfigMaps[key] = true
								end
							end
						end
					end
				end
			end
			for _, cm in ipairs(configmaps) do
				local cmKey = cm.metadata.namespace .. "/" .. cm.metadata.name
				local cmName = cm.metadata.name
				local cmNamespace = cm.metadata.namespace
				if not usedConfigMaps[cmKey] then
					check_event("失败", "[未使用] ConfigMap " .. cmNamespace .. "/" .. cmName .. " 没有被任何 Pod 使用", {namespace=cmNamespace, name=cmName})
				end
			end
			print("ConfigMap 未被使用检测完成")
		
```
