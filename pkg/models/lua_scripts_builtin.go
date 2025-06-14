package models

import (
	"github.com/weibaohui/k8m/pkg/constants"
)

// BuiltinLuaScriptsVersion 统一管理所有内置脚本的版本号
const BuiltinLuaScriptsVersion = "v1"

// BuiltinLuaScripts 内置检查脚本列表
var BuiltinLuaScripts = []InspectionLuaScript{
	{
		Name:        "Service Selector 检查",
		Description: "检查每个 Service 的 selector 是否有对应 Pod",
		Group:       "",
		Version:     "v1",
		Kind:        "Service",
		ScriptType:  constants.LuaScriptTypeBuiltin,
		ScriptCode:  "Builtin_Service_001",
		Script: `
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
						local pods, err = kubectl:GVK("", "v1", "Pod"):Cache(10):WithLabelSelector(labelSelector):List()
						local count = 0
						if not err and pods then
							for _, _ in pairs(pods) do count = count + 1 end
						end
						if count > 0 then
							check_event("正常", "Service " .. svc.metadata.name .. " selector 正常, 关联 Pod 数: " .. count, {name=svc.metadata.name, selector=labelSelector, podCount=count, namespace=svc.metadata.namespace})
						else
							check_event("失败", "Service " .. svc.metadata.name .. " selector " .. labelSelector .. " 应该至少一个pod, 但是现在没有", {name=svc.metadata.name, selector=labelSelector, namespace=svc.metadata.namespace})
						end
					end
				end
			else
				print("Service 列表获取失败: " .. tostring(err))
			end
		`,
	},

	{
		Name:        "ConfigMap 未被使用检测",
		Description: "检测所有未被 Pod 使用的 ConfigMap",
		Group:       "",
		Version:     "v1",
		Kind:        "ConfigMap",
		ScriptType:  constants.LuaScriptTypeBuiltin,
		ScriptCode:  "Builtin_ConfigMap_002",
		Script: `
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
		`,
	},
	{
		Name:        "ConfigMap 空数据检测",
		Description: "检测所有 data 和 binaryData 字段都为空的 ConfigMap",
		Group:       "",
		Version:     "v1",
		Kind:        "ConfigMap",
		ScriptType:  constants.LuaScriptTypeBuiltin,
		ScriptCode:  "Builtin_ConfigMap_003",
		Script: `
			local configmaps, err = kubectl:GVK("", "v1", "ConfigMap"):AllNamespace(""):List()
			if err then
				print("获取 ConfigMap 失败".. tostring(err))
				return
			end
			for _, cm in ipairs(configmaps) do
				local cmName = cm.metadata.name
				local cmNamespace = cm.metadata.namespace
				local isEmpty = true
				if cm.data then
					for k, v in pairs(cm.data) do
						isEmpty = false
						break
					end
				end
				if isEmpty and cm.binaryData then
					for k, v in pairs(cm.binaryData) do
						isEmpty = false
						break
					end
				end
				if isEmpty then
					check_event("失败", "[空数据] ConfigMap " .. cmNamespace .. "/" .. cmName .. " 的 data 和 binaryData 字段都为空", {namespace=cmNamespace, name=cmName})
				end
			end
			print("ConfigMap 空数据检测完成")
		`,
	},
	{
		Name:        "ConfigMap 超大检测",
		Description: "检测所有超过 1MB 的 ConfigMap",
		Group:       "",
		Version:     "v1",
		Kind:        "ConfigMap",
		ScriptType:  constants.LuaScriptTypeBuiltin,
		ScriptCode:  "Builtin_ConfigMap_004",
		Script: `
			local configmaps, err = kubectl:GVK("", "v1", "ConfigMap"):AllNamespace(""):List()
			if err then
				print( "获取 ConfigMap 失败".. tostring(err))
				return
			end
			for _, cm in ipairs(configmaps) do
				local cmName = cm.metadata.name
				local cmNamespace = cm.metadata.namespace
				local totalSize = 0
				if cm.data then
					for k, v in pairs(cm.data) do
						if type(v) == "string" then
							totalSize = totalSize + string.len(v)
						end
					end
				end
				if cm.binaryData then
					for k, v in pairs(cm.binaryData) do
						if type(v) == "string" then
							totalSize = totalSize + string.len(v)
						end
					end
				end
				local maxSize = 1024 * 1024
				if totalSize > maxSize then
					local sizeMB = string.format("%.2f", totalSize / (1024 * 1024))
					check_event("失败", "[超大] ConfigMap " .. cmNamespace .. "/" .. cmName .. " 大小为 " .. sizeMB .. "MB，超过 1MB 限制", {namespace=cmNamespace, name=cmName, size=sizeMB})
				end
			end
			print("ConfigMap 超大检测完成")
		`,
	},

	{
		Name:        "Deployment 配置检查",
		Description: "分析 Deployment 配置问题",
		Group:       "apps",
		Version:     "v1",
		Kind:        "Deployment",
		ScriptType:  constants.LuaScriptTypeBuiltin,
		ScriptCode:  "Builtin_Deployment_005",
		Script: `
			local deployments, err = kubectl:GVK("apps", "v1", "Deployment"):Cache(10):AllNamespace(""):List()
			if err then
				print( "获取 Deployment 失败".. tostring(err))
				return
			end
			local problemCount = 0
			for _, deployment in ipairs(deployments) do
				local deploymentName = deployment.metadata.name
				local deploymentNamespace = deployment.metadata.namespace
				local specReplicas = 0
				local statusReplicas = 0
				local readyReplicas = 0
				if deployment.spec and deployment.spec.replicas ~= nil then
					specReplicas = tonumber(deployment.spec.replicas) or 0
				end
				if deployment.status then
					if deployment.status.replicas ~= nil then
						statusReplicas = tonumber(deployment.status.replicas) or 0
					end
					if deployment.status.readyReplicas ~= nil then
						readyReplicas = tonumber(deployment.status.readyReplicas) or 0
					end
				end
				if specReplicas ~= readyReplicas then
					problemCount = problemCount + 1
					if statusReplicas > specReplicas then
						check_event("失败", "[副本数不匹配] Deployment " .. deploymentNamespace .. "/" .. deploymentName ..
							" 期望副本数: " .. specReplicas ..
							", 状态副本数: " .. statusReplicas ..
							", 就绪副本数: " .. readyReplicas ..
							" (状态字段尚未更新，缩容进行中)", {namespace=deploymentNamespace, name=deploymentName, specReplicas=specReplicas, statusReplicas=statusReplicas, readyReplicas=readyReplicas})
					else
						check_event("失败", "[副本数不足] Deployment " .. deploymentNamespace .. "/" .. deploymentName ..
							" 期望副本数: " .. specReplicas ..
							", 就绪副本数: " .. readyReplicas ..
							" (可能存在 Pod 启动失败或资源不足)", {namespace=deploymentNamespace, name=deploymentName, specReplicas=specReplicas, readyReplicas=readyReplicas})
					end
					if readyReplicas == 0 and specReplicas > 0 then
						check_event("失败", "没有就绪的副本，可能存在严重问题", {namespace=deploymentNamespace, name=deploymentName})
					elseif readyReplicas < specReplicas then
						local missingReplicas = specReplicas - readyReplicas
						check_event("失败", "缺少 " .. missingReplicas .. " 个副本，建议检查 Pod 状态和资源限制", {namespace=deploymentNamespace, name=deploymentName, missingReplicas=missingReplicas})
					end
				else
					check_event("正常", "[正常] Deployment " .. deploymentNamespace .. "/" .. deploymentName ..
						" 副本数正常: " .. specReplicas .. "/" .. readyReplicas, {namespace=deploymentNamespace, name=deploymentName, specReplicas=specReplicas, readyReplicas=readyReplicas})
				end
				if deployment.status and deployment.status.conditions then
					for _, condition in ipairs(deployment.status.conditions) do
						if condition.type == "Progressing" and condition.status == "False" then
							check_event("失败", "进度停滞: " .. (condition.reason or "未知原因") ..
								" - " .. (condition.message or "无详细信息"), {namespace=deploymentNamespace, name=deploymentName, reason=condition.reason, message=condition.message})
						elseif condition.type == "Available" and condition.status == "False" then
							check_event("失败", "不可用状态: " .. (condition.reason or "未知原因") ..
								" - " .. (condition.message or "无详细信息"), {namespace=deploymentNamespace, name=deploymentName, reason=condition.reason, message=condition.message})
						end
					end
				end
			end
		`,
	},
}
