# Pod 探针配置合规性检查 | 未配置存活/就绪探针

## 介绍

检查容器是否未配置 LivenessProbe/ReadinessProbe，避免 Pod 误杀或服务未就绪即接收流量。

## 信息

- ScriptCode: Builtin_Probe_001
- Kind: Pod
- Group: 
- Version: v1
- TimeoutSeconds: 120

## 代码

```lua

			local function list_all_pods()
				local pods, err = kubectl:GVK("", "v1", "Pod"):AllNamespace(""):List()
				if err then
					print("获取 Pod 失败: " .. tostring(err))
					return nil
				end
				if pods and pods.items then
					return pods.items
				end
				return pods
			end

			local pods = list_all_pods()
			if not pods then
				return
			end

			for _, pod in ipairs(pods) do
				local ns = pod.metadata and pod.metadata.namespace or ""
				local name = pod.metadata and pod.metadata.name or ""

				local containers = pod.spec and pod.spec.containers or nil
				if containers then
					for _, c in ipairs(containers) do
						local cName = c.name or ""
						if not c.livenessProbe then
							check_event(
								"失败",
								"Pod " .. ns .. "/" .. name .. " 容器 " .. cName .. " 未配置存活探针（LivenessProbe），可能导致异常无法自动恢复或误判",
								{ namespace = ns, name = name, container = cName, probe = "liveness" }
							)
						end
						if not c.readinessProbe then
							check_event(
								"失败",
								"Pod " .. ns .. "/" .. name .. " 容器 " .. cName .. " 未配置就绪探针（ReadinessProbe），可能导致未就绪即接收流量",
								{ namespace = ns, name = name, container = cName, probe = "readiness" }
							)
						end
					end
				end
			end

			print("Pod 探针缺失检查完成")
		
```
