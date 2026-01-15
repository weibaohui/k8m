# 镜像配置风险巡检 | 镜像拉取策略为 Always

## 介绍

检测 Pod 中容器 imagePullPolicy=Always，可能导致节点调度与启动延迟。

## 信息

- ScriptCode: Builtin_ImageRisk_003
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

				local function check_containers(containers, containerType)
					if not containers then
						return
					end
					for _, c in ipairs(containers) do
						local policy = c.imagePullPolicy or ""
						if policy == "Always" then
							local cName = c.name or ""
							local image = c.image or ""
							local typeDesc = containerType
							if typeDesc ~= "" then
								typeDesc = typeDesc .. " "
							end
							check_event(
								"失败",
								"Pod " .. ns .. "/" .. name .. " 的" .. typeDesc .. "容器 " .. cName .. " 镜像拉取策略为 Always，可能导致调度延迟: " .. image,
								{ namespace = ns, name = name, container = cName, image = image, image_pull_policy = policy, container_type = containerType }
							)
						end
					end
				end

				if pod.spec then
					check_containers(pod.spec.initContainers, "init")
					check_containers(pod.spec.containers, "")
				end
			end

			print("镜像拉取策略风险检查完成")
		
```
