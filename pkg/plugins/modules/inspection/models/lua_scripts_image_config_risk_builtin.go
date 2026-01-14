package models

import (
	"github.com/weibaohui/k8m/pkg/constants"
	"k8s.io/klog/v2"
)

var builtinLuaScriptsImageConfigRisk = []InspectionLuaScript{
	{
		Name:           "镜像配置风险巡检 | 容器使用 latest 标签",
		Description:    "检测 Pod 中容器镜像使用 latest（或未显式指定 tag）导致镜像版本不可控。",
		Group:          "",
		Version:        "v1",
		Kind:           "Pod",
		ScriptType:     constants.LuaScriptTypeBuiltin,
		ScriptCode:     "Builtin_ImageRisk_001",
		TimeoutSeconds: 120,
		Script: `
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
		`,
	},
	{
		Name:           "镜像配置风险巡检 | 镜像来自非信任仓库",
		Description:    "检测 Pod 中容器镜像来自公网匿名镜像仓库（如 Docker Hub/Quay）。",
		Group:          "",
		Version:        "v1",
		Kind:           "Pod",
		ScriptType:     constants.LuaScriptTypeBuiltin,
		ScriptCode:     "Builtin_ImageRisk_002",
		TimeoutSeconds: 120,
		Script: `
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
		`,
	},
	{
		Name:           "镜像配置风险巡检 | 镜像拉取策略为 Always",
		Description:    "检测 Pod 中容器 imagePullPolicy=Always，可能导致节点调度与启动延迟。",
		Group:          "",
		Version:        "v1",
		Kind:           "Pod",
		ScriptType:     constants.LuaScriptTypeBuiltin,
		ScriptCode:     "Builtin_ImageRisk_003",
		TimeoutSeconds: 120,
		Script: `
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
		`,
	},
}

// registerBuiltinImageConfigRiskLuaScripts 注册镜像配置风险巡检相关内置脚本。
func registerBuiltinImageConfigRiskLuaScripts() {
	BuiltinLuaScripts = append(BuiltinLuaScripts, builtinLuaScriptsImageConfigRisk...)
}

// init 初始化并注册镜像配置风险巡检内置脚本。
func init() {
	klog.V(6).Info("自动注册镜像配置风险巡检内置脚本")
	registerBuiltinImageConfigRiskLuaScripts()
}
