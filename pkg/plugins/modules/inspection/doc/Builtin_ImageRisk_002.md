# 镜像配置风险巡检 | 镜像来自非信任仓库

## 介绍

检测 Pod 中容器镜像来自公网匿名镜像仓库（如 Docker Hub/Quay）。

## 信息

- ScriptCode: Builtin_ImageRisk_002
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

			local function get_image_registry(image)
				if not image or image == "" then
					return ""
				end
				local first = string.match(image, "^([^/]+)/")
				if not first then
					return "docker.io"
				end
				if string.find(first, "%.") or string.find(first, ":") or first == "localhost" then
					return first
				end
				return "docker.io"
			end

			local untrusted_registries = {
				["docker.io"] = true,
				["index.docker.io"] = true,
				["registry-1.docker.io"] = true,
				["quay.io"] = true,
			}

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
						local image = c.image or ""
						local registry = get_image_registry(image)
						if registry ~= "" and untrusted_registries[registry] then
							local cName = c.name or ""
							local typeDesc = containerType
							if typeDesc ~= "" then
								typeDesc = typeDesc .. " "
							end
							check_event(
								"失败",
								"Pod " .. ns .. "/" .. name .. " 的" .. typeDesc .. "容器 " .. cName .. " 镜像来自非信任仓库: " .. registry .. "，镜像=" .. image,
								{ namespace = ns, name = name, container = cName, image = image, registry = registry, container_type = containerType }
							)
						end
					end
				end

				if pod.spec then
					check_containers(pod.spec.initContainers, "init")
					check_containers(pod.spec.containers, "")
				end
			end

			print("镜像仓库可信性风险检查完成")
		
```
