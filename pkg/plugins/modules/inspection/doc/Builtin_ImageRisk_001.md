# 镜像配置风险巡检 | 容器使用 latest 标签

## 介绍

检测 Pod 中容器镜像使用 latest（或未显式指定 tag）导致镜像版本不可控。

## 信息

- ScriptCode: Builtin_ImageRisk_001
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

			local function get_image_tag(image)
				if not image or image == "" then
					return ""
				end
				if string.find(image, "@") then
					return ""
				end
				local last = string.match(image, ".*/([^/]+)$") or image
				local tag = string.match(last, ":(.+)$")
				if tag then
					return tag
				end
				return ""
			end

			local function is_latest_image(image)
				if not image or image == "" then
					return false
				end
				if string.find(image, "@") then
					return false
				end
				local tag = get_image_tag(image)
				if tag == "" or tag == "latest" then
					return true
				end
				return false
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
						local image = c.image or ""
						if is_latest_image(image) then
							local cName = c.name or ""
							local typeDesc = containerType
							if typeDesc ~= "" then
								typeDesc = typeDesc .. " "
							end
							check_event(
								"失败",
								"Pod " .. ns .. "/" .. name .. " 的" .. typeDesc .. "容器 " .. cName .. " 使用 latest 标签（镜像版本不可控）: " .. image,
								{ namespace = ns, name = name, container = cName, image = image, container_type = containerType }
							)
						end
					end
				end

				if pod.spec then
					check_containers(pod.spec.initContainers, "init")
					check_containers(pod.spec.containers, "")
				end
			end

			print("镜像 latest 标签风险检查完成")
		
```
