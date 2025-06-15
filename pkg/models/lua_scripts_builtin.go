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
						if count > 0 then
							check_event("正常", "Service " .. svc.metadata.name .. " selector 正常, 关联 Pod 数: " .. count .. "spec.selector定义" .. doc, {name=svc.metadata.name, selector=labelSelector, podCount=count, namespace=svc.metadata.namespace})
						else
							check_event("失败", "Service " .. svc.metadata.name .. " selector " .. labelSelector .. " 应该至少一个pod, 但是现在没有。" .. "spec.selector定义" .. doc, {name=svc.metadata.name, selector=labelSelector, namespace=svc.metadata.namespace})
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

			local doc, err = kubectl:GVK("apps", "v1", "Deployment"):Cache(10):Doc("spec.replicas")
			if err then
				print( "获取 Deployment Doc 失败".. tostring(err))
				return
			end
			print("Deployment Doc 获取成功: " .. doc)
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
	{
		Name:        "CronJob 合规性检查",
		Description: "检查 CronJob 是否被挂起、调度表达式是否合法、startingDeadlineSeconds 是否为负数",
		Group:       "",
		Version:     "v1",
		Kind:        "CronJob",
		ScriptType:  constants.LuaScriptTypeBuiltin,
		ScriptCode:  "Builtin_CronJob_006",
		Script: `
			local cron = require("cron")
			local cronjobs, err = kubectl:GVK("batch", "v1", "CronJob"):AllNamespace(""):List()
			if err then
				print("获取 CronJob 失败: " .. tostring(err))
				return
			end
			local doc_suspend, _ = kubectl:GVK("batch", "v1", "CronJob"):Doc("spec.suspend")
			local doc_schedule, _ = kubectl:GVK("batch", "v1", "CronJob"):Doc("spec.schedule")
			local doc_deadline, _ = kubectl:GVK("batch", "v1", "CronJob"):Doc("spec.startingDeadlineSeconds")
			for _, cj in ipairs(cronjobs) do
				local ns = cj.metadata and cj.metadata.namespace or "default"
				local name = cj.metadata and cj.metadata.name or ""
				-- 检查挂起
				if cj.spec and cj.spec.suspend == true then
					check_event("失败", "CronJob " .. name .. " 已被挂起", {namespace=ns, name=name, doc=doc_suspend})
				end
				-- 检查 startingDeadlineSeconds
				if cj.spec and cj.spec.startingDeadlineSeconds ~= nil then
					if tonumber(cj.spec.startingDeadlineSeconds) < 0 then
						check_event("失败", "CronJob " .. name .. " 的 startingDeadlineSeconds 为负数", {namespace=ns, name=name, value=cj.spec.startingDeadlineSeconds, doc=doc_deadline})
					end
				end
			end
			print("CronJob 合规性检查完成")
		`,
	},
	{
		Name:        "Gateway 合规性检查",
		Description: "检查 Gateway 关联的 GatewayClass 是否存在，以及 Gateway 状态是否被接受",
		Group:       "gateway.networking.k8s.io",
		Version:     "v1",
		Kind:        "Gateway",
		ScriptType:  constants.LuaScriptTypeBuiltin,
		ScriptCode:  "Builtin_Gateway_007",
		Script: `
			local gateways, err = kubectl:GVK("gateway.networking.k8s.io", "v1", "Gateway"):AllNamespace(""):List()
			if err then
				print("获取 Gateway 失败: " .. tostring(err))
				return
			end
			for _, gtw in ipairs(gateways) do
				local ns = gtw.metadata and gtw.metadata.namespace or "default"
				local name = gtw.metadata and gtw.metadata.name or ""
				local className = gtw.spec and gtw.spec.gatewayClassName or nil
				local classExists = false
				if className then
					local gtwclass, err = kubectl:GVK("gateway.networking.k8s.io", "v1", "GatewayClass"):Name(className):Get()
					if not err and gtwclass then
						classExists = true
					end
				end
				if not classExists then
					check_event("失败", "Gateway 使用的 GatewayClass " .. tostring(className) .. " 不存在", {namespace=ns, name=name, gatewayClassName=className})
				end
				-- 检查第一个 Condition 状态
				if gtw.status and gtw.status.conditions and #gtw.status.conditions > 0 then
					local cond = gtw.status.conditions[1]
					if cond.status ~= "True" then
						check_event("失败", "Gateway '" .. ns .. "/" .. name .. "' 未被接受, Message: '" .. (cond.message or "") .. "'", {namespace=ns, name=name, message=cond.message})
					end
				end
			end
			print("Gateway 合规性检查完成")
		`,
	},
	{
		Name:        "GatewayClass 合规性检查",
		Description: "检查 GatewayClass 的第一个 Condition 状态是否为 True，否则报告未被接受及 message。",
		Group:       "gateway.networking.k8s.io",
		Version:     "v1",
		Kind:        "GatewayClass",
		ScriptType:  constants.LuaScriptTypeBuiltin,
		ScriptCode:  "Builtin_GatewayClass_008",
		Script: `
			local gatewayclasses, err = kubectl:GVK("gateway.networking.k8s.io", "v1", "GatewayClass"):AllNamespace(""):List()
			if err then
				print("获取 GatewayClass 失败: " .. tostring(err))
				return
			end
			for _, gc in ipairs(gatewayclasses) do
				local name = gc.metadata and gc.metadata.name or ""
				if gc.status and gc.status.conditions and #gc.status.conditions > 0 then
					local cond = gc.status.conditions[1]
					if cond.status ~= "True" then
						check_event("失败", "GatewayClass '" .. name .. "' 未被接受, Message: '" .. (cond.message or "") .. "'", {name=name, message=cond.message})
					end
				end
			end
			print("GatewayClass 合规性检查完成")
		`,
	},
	{
		Name:        "HPA Condition 检查",
		Description: "检查 HorizontalPodAutoscaler 的 Condition 状态，ScalingLimited 为 True 或其他 Condition 为 False 时报警。",
		Group:       "autoscaling",
		Version:     "v2",
		Kind:        "HorizontalPodAutoscaler",
		ScriptType:  constants.LuaScriptTypeBuiltin,
		ScriptCode:  "Builtin_HPA_Condition_009",
		Script: `
			local hpas, err = kubectl:GVK("autoscaling", "v2", "HorizontalPodAutoscaler"):AllNamespace(""):List()
			if err then print("获取 HPA 失败: " .. tostring(err)) return end
			for _, hpa in ipairs(hpas) do
				if hpa.status and hpa.status.conditions then
					for _, cond in ipairs(hpa.status.conditions) do
						if cond.type == "ScalingLimited" and cond.status == "True" then
							check_event("失败", cond.message or "ScalingLimited condition True", {namespace=hpa.metadata.namespace, name=hpa.metadata.name, type=cond.type})
						elseif cond.status == "False" then
							check_event("失败", cond.message or (cond.type .. " condition False"), {namespace=hpa.metadata.namespace, name=hpa.metadata.name, type=cond.type})
						end
					end
				end
			end
			print("HPA Condition 检查完成")
		`,
	},
	{
		Name:        "HPA ScaleTargetRef 存在性检查",
		Description: "检查 HorizontalPodAutoscaler 的 ScaleTargetRef 指向的对象是否存在。",
		Group:       "autoscaling",
		Version:     "v2",
		Kind:        "HorizontalPodAutoscaler",
		ScriptType:  constants.LuaScriptTypeBuiltin,
		ScriptCode:  "Builtin_HPA_ScaleTargetRef_010",
		Script: `
			local hpas, err = kubectl:GVK("autoscaling", "v2", "HorizontalPodAutoscaler"):AllNamespace(""):List()
			if err then print("获取 HPA 失败: " .. tostring(err)) return end
			for _, hpa in ipairs(hpas) do
				if hpa.spec and hpa.spec.scaleTargetRef then
					local ref = hpa.spec.scaleTargetRef
					local exists = false
					if ref.kind == "Deployment" then
						exists = kubectl:GVK("apps", "v1", "Deployment"):Namespace(hpa.metadata.namespace):Name(ref.name):Exists()
					elseif ref.kind == "ReplicaSet" then
						exists = kubectl:GVK("apps", "v1", "ReplicaSet"):Namespace(hpa.metadata.namespace):Name(ref.name):Exists()
					elseif ref.kind == "StatefulSet" then
						exists = kubectl:GVK("apps", "v1", "StatefulSet"):Namespace(hpa.metadata.namespace):Name(ref.name):Exists()
					elseif ref.kind == "ReplicationController" then
						exists = kubectl:GVK("", "v1", "ReplicationController"):Namespace(hpa.metadata.namespace):Name(ref.name):Exists()
					else
						check_event("失败", "HorizontalPodAutoscaler 使用了不支持的 ScaleTargetRef Kind: " .. tostring(ref.kind), {namespace=hpa.metadata.namespace, name=hpa.metadata.name, kind=ref.kind})
					end
					if not exists then
						check_event("失败", "HorizontalPodAutoscaler 的 ScaleTargetRef " .. ref.kind .. "/" .. ref.name .. " 不存在", {namespace=hpa.metadata.namespace, name=hpa.metadata.name, kind=ref.kind, refname=ref.name})
					end
				end
			end
			print("HPA ScaleTargetRef 存在性检查完成")
		`,
	},
	{
		Name:        "HPA 资源配置检查",
		Description: "检查 HPA 关联对象的 Pod 模板中所有容器是否配置了 requests 和 limits。",
		Group:       "autoscaling",
		Version:     "v2",
		Kind:        "HorizontalPodAutoscaler",
		ScriptType:  constants.LuaScriptTypeBuiltin,
		ScriptCode:  "Builtin_HPA_Resource_011",
		Script: `
			local hpas, err = kubectl:GVK("autoscaling", "v2", "HorizontalPodAutoscaler"):AllNamespace(""):List()
			if err then print("获取 HPA 失败: " .. tostring(err)) return end
			for _, hpa in ipairs(hpas) do
				if hpa.spec and hpa.spec.scaleTargetRef then
					local ref = hpa.spec.scaleTargetRef
					local gvk_map = {
						Deployment = {group="apps", version="v1", kind="Deployment"},
						ReplicaSet = {group="apps", version="v1", kind="ReplicaSet"},
						StatefulSet = {group="apps", version="v1", kind="StatefulSet"},
						ReplicationController = {group="", version="v1", kind="ReplicationController"},
					}
					local gvk = gvk_map[ref.kind]
					if gvk then
						local target, err = kubectl:GVK(gvk.group, gvk.version, gvk.kind):Namespace(hpa.metadata.namespace):Name(ref.name):Get()
						if not err and target and target.spec and target.spec.template and target.spec.template.spec and target.spec.template.spec.containers then
							local containers = target.spec.template.spec.containers
							local all_ok = true
							for _, c in ipairs(containers) do
								if not c.resources or not c.resources.requests or not c.resources.limits then
									all_ok = false
									check_event("失败", ref.kind .. " " .. hpa.metadata.namespace .. "/" .. ref.name .. " 的容器未配置 requests 或 limits", {namespace=hpa.metadata.namespace, name=hpa.metadata.name, kind=ref.kind, refname=ref.name, container=c.name})
								end
							end
							if all_ok then
								check_event("正常", ref.kind .. " " .. hpa.metadata.namespace .. "/" .. ref.name .. " 的所有容器资源配置齐全", {namespace=hpa.metadata.namespace, name=hpa.metadata.name, kind=ref.kind, refname=ref.name})
							end
						end
					end
				end
			end
			print("HPA 资源配置检查完成")
		`,
	},
}
