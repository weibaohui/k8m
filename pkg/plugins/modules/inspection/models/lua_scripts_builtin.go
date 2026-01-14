package models

import (
	"github.com/weibaohui/k8m/pkg/constants"
)

// BuiltinLuaScriptsVersion 统一管理所有内置脚本的版本号
const BuiltinLuaScriptsVersion = "v2"

// BuiltinLuaScripts 内置检查脚本列表
var BuiltinLuaScripts = []InspectionLuaScript{
	{
		Name:           "Service Selector 检查",
		Description:    "检查每个 Service 的 selector 是否有对应 Pod",
		Group:          "",
		Version:        "v1",
		Kind:           "Service",
		ScriptType:     constants.LuaScriptTypeBuiltin,
		ScriptCode:     "Builtin_Service_001",
		TimeoutSeconds: 30, // Service检查相对简单，30秒足够
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
						if count == 0 then
							check_event("失败", "Service " .. svc.metadata.name .. " selector " .. labelSelector .. " 应该至少一个pod, 但是现在没有。" .. "spec.selector定义" .. doc, {name=svc.metadata.name, selector=labelSelector, namespace=svc.metadata.namespace})
						end
					end
				end
			else
				print("Service 列表获取失败: " .. tostring(err))
			end
			print("Service Selector 检查完成")
		`,
	},

	{
		Name:           "ConfigMap 未被使用检测",
		Description:    "检测所有未被 Pod 使用的 ConfigMap",
		Group:          "",
		Version:        "v1",
		Kind:           "ConfigMap",
		ScriptType:     constants.LuaScriptTypeBuiltin,
		ScriptCode:     "Builtin_ConfigMap_002",
		TimeoutSeconds: 90, // 需要遍历所有Pod和ConfigMap，时间较长
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
		Name:           "ConfigMap 空数据检测",
		Description:    "检测所有 data 和 binaryData 字段都为空的 ConfigMap",
		Group:          "",
		Version:        "v1",
		Kind:           "ConfigMap",
		ScriptType:     constants.LuaScriptTypeBuiltin,
		ScriptCode:     "Builtin_ConfigMap_003",
		TimeoutSeconds: 30, // 简单的数据检查，30秒足够
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
		Name:           "ConfigMap 超大检测",
		Description:    "检测所有超过 1MB 的 ConfigMap",
		Group:          "",
		Version:        "v1",
		Kind:           "ConfigMap",
		ScriptType:     constants.LuaScriptTypeBuiltin,
		ScriptCode:     "Builtin_ConfigMap_004",
		TimeoutSeconds: 45, // 需要计算数据大小，稍微复杂一些
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
		Name:           "Deployment 配置检查",
		Description:    "分析 Deployment 配置问题",
		Group:          "apps",
		Version:        "v1",
		Kind:           "Deployment",
		ScriptType:     constants.LuaScriptTypeBuiltin,
		ScriptCode:     "Builtin_Deployment_005",
		TimeoutSeconds: 60, // 需要检查状态和条件，使用默认60秒
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
		Name:           "CronJob 合规性检查",
		Description:    "检查 CronJob 是否被挂起、调度表达式是否合法、startingDeadlineSeconds 是否为负数",
		Group:          "",
		Version:        "v1",
		Kind:           "CronJob",
		ScriptType:     constants.LuaScriptTypeBuiltin,
		ScriptCode:     "Builtin_CronJob_006",
		TimeoutSeconds: 45, // 包含复杂的Cron表达式验证逻辑
		Script: `
			-- 内置 Cron 表达式基本校验（Kubernetes 使用标准 5 字段）
			local function split_fields(expr)
				local fields = {}
				for token in string.gmatch(expr or "", "%S+") do table.insert(fields, token) end
				return fields
			end
			local function validate_part(part, min, max, allow_names)
				if part == "*" then return true end
				local step = string.match(part, "^%*/(%d+)$")
				if step then return tonumber(step) and tonumber(step) >= 1 end
				local a,b = string.match(part, "^(%d+)%-(%d+)$")
				if a and b then a=tonumber(a); b=tonumber(b); return a and b and a>=min and b<=max and a<=b end
				local num = tonumber(part)
				if num and num>=min and num<=max then return true end
				if allow_names and string.match(part, "^[A-Za-z]+$") then return true end
				return false
			end
			local function validate_field(field, min, max, allow_names)
				for part in string.gmatch(field, "[^,]+") do
					if not validate_part(part, min, max, allow_names) then return false end
				end
				return true
			end
			local function is_valid_cron(expr)
				if not expr or expr == "" then return false, "表达式为空" end
				if string.match(expr, "^@%w+$") then return true end -- 支持 @yearly 等描述符
				local f = split_fields(expr)
				if #f ~= 5 then return false, "字段数不是5" end
				if not validate_field(f[1], 0, 59, false) then return false, "分钟字段非法" end
				if not validate_field(f[2], 0, 23, false) then return false, "小时字段非法" end
				if not validate_field(f[3], 1, 31, false) then return false, "日字段非法" end
				if not validate_field(f[4], 1, 12, true) then return false, "月字段非法" end
				if not validate_field(f[5], 0, 7, true) then return false, "周字段非法" end
				return true
			end
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
				-- 检查调度表达式合法性（5 字段）
				if cj.spec and cj.spec.schedule ~= nil then
					local ok, reason = is_valid_cron(cj.spec.schedule)
					if not ok then
						check_event("失败", "CronJob " .. name .. " 的调度表达式非法: " .. tostring(reason), {namespace=ns, name=name, value=cj.spec.schedule, doc=doc_schedule})
					end
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
		Name:           "Gateway 合规性检查",
		Description:    "检查 Gateway 关联的 GatewayClass 是否存在，以及 Gateway 状态是否被接受",
		Group:          "gateway.networking.k8s.io",
		Version:        "v1",
		Kind:           "Gateway",
		ScriptType:     constants.LuaScriptTypeBuiltin,
		ScriptCode:     "Builtin_Gateway_007",
		TimeoutSeconds: 45, // 需要检查GatewayClass存在性和状态
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
		Name:           "GatewayClass 合规性检查",
		Description:    "检查 GatewayClass 的第一个 Condition 状态是否为 True，否则报告未被接受及 message。",
		Group:          "gateway.networking.k8s.io",
		Version:        "v1",
		Kind:           "GatewayClass",
		ScriptType:     constants.LuaScriptTypeBuiltin,
		ScriptCode:     "Builtin_GatewayClass_008",
		TimeoutSeconds: 45, // 需要检查Gateway引用和状态
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
		Name:           "HPA Condition 检查",
		Description:    "检查 HorizontalPodAutoscaler 的 Condition 状态，ScalingLimited 为 True 或其他 Condition 为 False 时报警。",
		Group:          "autoscaling",
		Version:        "v2",
		Kind:           "HorizontalPodAutoscaler",
		ScriptType:     constants.LuaScriptTypeBuiltin,
		ScriptCode:     "Builtin_HPA_Condition_009",
		TimeoutSeconds: 45, // HPA状态检查，需要一定时间
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
		Name:           "HPA ScaleTargetRef 存在性检查",
		Description:    "检查 HorizontalPodAutoscaler 的 ScaleTargetRef 指向的对象是否存在。",
		Group:          "autoscaling",
		Version:        "v2",
		Kind:           "HorizontalPodAutoscaler",
		ScriptType:     constants.LuaScriptTypeBuiltin,
		ScriptCode:     "Builtin_HPA_ScaleTargetRef_010",
		TimeoutSeconds: 60, // 需要检查多种资源类型的存在性
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
		Name:           "HPA 资源配置检查",
		Description:    "检查 HPA 关联对象的 Pod 模板中所有容器是否配置了 requests 和 limits。",
		Group:          "autoscaling",
		Version:        "v2",
		Kind:           "HorizontalPodAutoscaler",
		ScriptType:     constants.LuaScriptTypeBuiltin,
		ScriptCode:     "Builtin_HPA_Resource_011",
		TimeoutSeconds: 75, // 需要检查HPA和关联的Deployment/StatefulSet等资源
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
						end
					end
				end
			end
			print("HPA 资源配置检查完成")
		`,
	}, {
		Name:           "HTTPRoute Backend Service 存在性与端口检查",
		Description:    "检查 HTTPRoute 所引用的后端 Service 是否存在，以及端口是否匹配 Service 的端口。",
		Group:          "gateway.networking.k8s.io",
		Version:        "v1",
		Kind:           "HTTPRoute",
		ScriptType:     constants.LuaScriptTypeBuiltin,
		ScriptCode:     "Builtin_HTTPRoute_Backend_012",
		TimeoutSeconds: 60, // 需要检查HTTPRoute和Service的存在性及端口匹配
		Script: `
			local httproutes, err = kubectl:GVK("gateway.networking.k8s.io", "v1", "HTTPRoute"):AllNamespace(""):List()
			if err then print("获取 HTTPRoute 失败: " .. tostring(err)) return end
			for _, route in ipairs(httproutes) do
				if route.spec and route.spec.rules then
					for _, rule in ipairs(route.spec.rules) do
						if rule.backendRefs then
							for _, backend in ipairs(rule.backendRefs) do
								local svc, err = kubectl:GVK("", "v1", "Service"):Namespace(route.metadata.namespace):Name(backend.name):Get()
								if err or not svc then
									check_event("失败", "HTTPRoute 使用的 Service '" .. route.metadata.namespace .. "/" .. backend.name .. "' 不存在", {namespace=route.metadata.namespace, name=backend.name})
								else
									local portMatch = false
									if svc.spec and svc.spec.ports and backend.port then
										for _, svcPort in ipairs(svc.spec.ports) do
											if svcPort.port == backend.port then portMatch = true end
										end
									end
									if not portMatch then
										check_event("失败", "HTTPRoute 的后端 Service '" .. backend.name .. "' 使用端口 '" .. tostring(backend.port) .. "'，但 Service 未配置该端口", {namespace=route.metadata.namespace, name=backend.name, port=backend.port})
									end
								end
							end
						end
					end
				end
			end
			print("HTTPRoute Backend Service 检查完成")
		`,
	}, {
		Name:           "HTTPRoute Backend Service 存在性与端口检查",
		Description:    "检查 HTTPRoute 所引用的后端 Service 是否存在，以及端口是否匹配 Service 的端口。",
		Group:          "gateway.networking.k8s.io",
		Version:        "v1",
		Kind:           "HTTPRoute",
		ScriptType:     constants.LuaScriptTypeBuiltin,
		ScriptCode:     "Builtin_HTTPRoute_Backend_013",
		TimeoutSeconds: 60, // 需要检查HTTPRoute和Service的存在性及端口匹配
		Script: `
			local httproutes, err = kubectl:GVK("gateway.networking.k8s.io", "v1", "HTTPRoute"):AllNamespace(""):List()
			if err then print("获取 HTTPRoute 失败: " .. tostring(err)) return end
			for _, route in ipairs(httproutes) do
				if route.spec and route.spec.rules then
					for _, rule in ipairs(route.spec.rules) do
						if rule.backendRefs then
							for _, backend in ipairs(rule.backendRefs) do
								local svc, err = kubectl:GVK("", "v1", "Service"):Namespace(route.metadata.namespace):Name(backend.name):Get()
								if err or not svc then
									check_event("失败", "HTTPRoute 使用的 Service '" .. route.metadata.namespace .. "/" .. backend.name .. "' 不存在", {namespace=route.metadata.namespace, name=backend.name})
								else
									local portMatch = false
									if svc.spec and svc.spec.ports and backend.port then
										for _, svcPort in ipairs(svc.spec.ports) do
											if svcPort.port == backend.port then portMatch = true end
										end
									end
									if not portMatch then
										check_event("失败", "HTTPRoute 的后端 Service '" .. backend.name .. "' 使用端口 '" .. tostring(backend.port) .. "'，但 Service 未配置该端口", {namespace=route.metadata.namespace, name=backend.name, port=backend.port})
									end
								end
							end
						end
					end
				end
			end
			print("HTTPRoute Backend Service 检查完成")
		`,
	}, {
		Name:           "HTTPRoute Gateway 存在性与命名空间策略检查",
		Description:    "检查 HTTPRoute 所引用的 Gateway 是否存在，以及 Gateway 的 AllowedRoutes 策略是否允许该 HTTPRoute。",
		Group:          "gateway.networking.k8s.io",
		Version:        "v1",
		Kind:           "HTTPRoute",
		ScriptType:     constants.LuaScriptTypeBuiltin,
		ScriptCode:     "Builtin_HTTPRoute_Gateway_014",
		TimeoutSeconds: 75, // 需要检查HTTPRoute、Gateway存在性和复杂的命名空间策略
		Script: `
			local httproutes, err = kubectl:GVK("gateway.networking.k8s.io", "v1", "HTTPRoute"):AllNamespace(""):List()
			if err then print("获取 HTTPRoute 失败: " .. tostring(err)) return end
			for _, route in ipairs(httproutes) do
				if route.spec and route.spec.parentRefs then
					for _, gtwref in ipairs(route.spec.parentRefs) do
						local ns = route.metadata.namespace
						if gtwref.namespace then ns = gtwref.namespace end
						local gtw, err = kubectl:GVK("gateway.networking.k8s.io", "v1", "Gateway"):Namespace(ns):Name(gtwref.name):Get()
						if err or not gtw then
							check_event("失败", "HTTPRoute 使用的 Gateway '" .. ns .. "/" .. gtwref.name .. "' 不存在", {namespace=ns, name=gtwref.name})
						else
							if gtw.spec and gtw.spec.listeners then
								for _, listener in ipairs(gtw.spec.listeners) do
									if listener.allowedRoutes and listener.allowedRoutes.namespaces and listener.allowedRoutes.namespaces.from then
										local allow = listener.allowedRoutes.namespaces.from
										if allow == "Same" and route.metadata.namespace ~= gtw.metadata.namespace then
											check_event("失败", "HTTPRoute '" .. route.metadata.namespace .. "/" .. route.metadata.name .. "' 与 Gateway '" .. gtw.metadata.namespace .. "/" .. gtw.metadata.name .. "' 不在同一命名空间，且 Gateway 只允许同命名空间 HTTPRoute", {route_ns=route.metadata.namespace, route_name=route.metadata.name, gtw_ns=gtw.metadata.namespace, gtw_name=gtw.metadata.name})
										elseif allow == "Selector" and listener.allowedRoutes.namespaces.selector and listener.allowedRoutes.namespaces.selector.matchLabels then
											local match = false
											for k, v in pairs(listener.allowedRoutes.namespaces.selector.matchLabels) do
												if route.metadata.labels and route.metadata.labels[k] == v then match = true end
											end
											if not match then
												check_event("失败", "HTTPRoute '" .. route.metadata.namespace .. "/" .. route.metadata.name .. "' 的标签与 Gateway '" .. gtw.metadata.namespace .. "/" .. gtw.metadata.name .. "' 的 Selector 不匹配", {route_ns=route.metadata.namespace, route_name=route.metadata.name, gtw_ns=gtw.metadata.namespace, gtw_name=gtw.metadata.name})
											end
										end
									end
								end
							end
						end
					end
				end
			end
			print("HTTPRoute Gateway 检查完成")
		`,
	},
	{
		Name:           "Ingress 合规性检查",
		Description:    "检查 Ingress 是否指定 IngressClass，引用的 IngressClass/Service/Secret 是否存在。",
		Group:          "networking",
		Version:        "v1",
		Kind:           "Ingress",
		ScriptType:     constants.LuaScriptTypeBuiltin,
		ScriptCode:     "Builtin_Ingress_015",
		TimeoutSeconds: 75, // 需要检查Ingress、IngressClass、Service和Secret的存在性
		Script: `
			local ingresses, err = kubectl:GVK("networking.k8s.io", "v1", "Ingress"):AllNamespace(""):List()
			if err then print("获取 Ingress 失败: " .. tostring(err)) return end
			for _, ing in ipairs(ingresses) do
				local ingressClassName = ing.spec and ing.spec.ingressClassName or nil
				if not ingressClassName and ing.metadata and ing.metadata.annotations then
					ingressClassName = ing.metadata.annotations["kubernetes.io/ingress.class"]
				end
				if not ingressClassName or ingressClassName == "" then
					check_event("失败", "Ingress " .. ing.metadata.namespace .. "/" .. ing.metadata.name .. " 未指定 IngressClass", {namespace=ing.metadata.namespace, name=ing.metadata.name})
				else
					local ic, err = kubectl:GVK("networking.k8s.io", "v1", "IngressClass"):Name(ingressClassName):Get()
					if err or not ic then
						check_event("失败", "Ingress 使用的 IngressClass '" .. ingressClassName .. "' 不存在", {namespace=ing.metadata.namespace, name=ing.metadata.name, ingressClass=ingressClassName})
					end
				end
				if ing.spec and ing.spec.rules then
					for _, rule in ipairs(ing.spec.rules) do
						if rule.http and rule.http.paths then
							for _, path in ipairs(rule.http.paths) do
								if path.backend and path.backend.service and path.backend.service.name then
									local svc, err = kubectl:GVK("", "v1", "Service"):Namespace(ing.metadata.namespace):Name(path.backend.service.name):Get()
									if err or not svc then
										check_event("失败", "Ingress 使用的 Service '" .. ing.metadata.namespace .. "/" .. path.backend.service.name .. "' 不存在", {namespace=ing.metadata.namespace, name=path.backend.service.name})
									end
								end
							end
						end
					end
				end
				if ing.spec and ing.spec.tls then
					for _, tls in ipairs(ing.spec.tls) do
						if tls.secretName then
							local sec, err = kubectl:GVK("", "v1", "Secret"):Namespace(ing.metadata.namespace):Name(tls.secretName):Get()
							if err or not sec then
								check_event("失败", "Ingress 使用的 TLS Secret '" .. ing.metadata.namespace .. "/" .. tls.secretName .. "' 不存在", {namespace=ing.metadata.namespace, name=tls.secretName})
							end
						end
					end
				end
			end
			print("Ingress 合规性检查完成")
		`,
	},
	{
		Name:           "Job 合规性检查",
		Description:    "检查 Job 是否被挂起（suspend）以及是否有失败（status.failed > 0）",
		Group:          "batch",
		Version:        "v1",
		Kind:           "Job",
		ScriptType:     constants.LuaScriptTypeBuiltin,
		ScriptCode:     "Builtin_Job_016",
		TimeoutSeconds: 45, // Job状态检查相对简单
		Script: `
			local jobs, err = kubectl:GVK("batch", "v1", "Job"):AllNamespace(""):List()
			if err then print("获取 Job 失败: " .. tostring(err)) return end
			for _, job in ipairs(jobs) do
				if job.spec and job.spec.suspend == true then
					check_event("失败", "Job " .. job.metadata.name .. " 已被挂起", {namespace=job.metadata.namespace, name=job.metadata.name})
				end
				if job.status and job.status.failed and job.status.failed > 0 then
					check_event("失败", "Job " .. job.metadata.name .. " 有失败记录 (failed=" .. tostring(job.status.failed) .. ")", {namespace=job.metadata.namespace, name=job.metadata.name, failed=job.status.failed})
				end
			end
			print("Job 合规性检查完成")
		`,
	},
	{
		Name:           "MutatingWebhookConfiguration 合规性检查",
		Description:    "检查 MutatingWebhookConfiguration 的 webhook 指向的 Service 是否存在、是否有活跃 Pod、Pod 状态。",
		Group:          "admissionregistration.k8s.io",
		Version:        "v1",
		Kind:           "MutatingWebhookConfiguration",
		ScriptType:     constants.LuaScriptTypeBuiltin,
		ScriptCode:     "Builtin_MutatingWebhook_017",
		TimeoutSeconds: 90, // 需要检查Service和Pod状态，较为复杂
		Script: `
			local mwcs, err = kubectl:GVK("admissionregistration.k8s.io", "v1", "MutatingWebhookConfiguration"):AllNamespace(""):List()
			if err then print("获取 MutatingWebhookConfiguration 失败: " .. tostring(err)) return end
			for _, mwc in ipairs(mwcs) do
				if mwc.webhooks then
					for _, webhook in ipairs(mwc.webhooks) do
						if webhook.clientConfig and webhook.clientConfig.service then
							local svc = webhook.clientConfig.service
							local service, err = kubectl:GVK("", "v1", "Service"):Namespace(svc.namespace):Name(svc.name):Get()
							if err or not service then
								check_event("失败", "MutatingWebhook " .. webhook.name .. " 指向的 Service '" .. svc.namespace .. "/" .. svc.name .. "' 不存在", {namespace=svc.namespace, name=svc.name, webhook=webhook.name})
							else
								if service.spec and service.spec.selector and next(service.spec.selector) ~= nil then
									local selector = ""
									for k, v in pairs(service.spec.selector) do
										if selector ~= "" then selector = selector .. "," end
										selector = selector .. k .. "=" .. v
									end
									local pods, err = kubectl:GVK("", "v1", "Pod"):Namespace(svc.namespace):WithLabelSelector(selector):List()
									if not err and pods and #pods.items == 0 then
										check_event("失败", "MutatingWebhook " .. webhook.name .. " 指向的 Service '" .. svc.namespace .. "/" .. svc.name .. "' 没有活跃 Pod", {namespace=svc.namespace, name=svc.name, webhook=webhook.name})
									end
									if pods and pods.items then
										for _, pod in ipairs(pods.items) do
											if pod.status and pod.status.phase ~= "Running" then
												check_event("失败", "MutatingWebhook " .. webhook.name .. " 指向的 Pod '" .. pod.metadata.name .. "' 状态为 " .. (pod.status.phase or "未知") , {namespace=svc.namespace, name=svc.name, webhook=webhook.name, pod=pod.metadata.name, phase=pod.status.phase})
											end
										end
									end
								end
							end
						end
					end
				end
			end
			print("MutatingWebhookConfiguration 合规性检查完成")
		`,
	},
	{
		Name:           "NetworkPolicy 合规性检查",
		Description:    "检查 NetworkPolicy 是否允许所有 Pod，或未作用于任何 Pod。",
		Group:          "networking",
		Version:        "v1",
		Kind:           "NetworkPolicy",
		ScriptType:     constants.LuaScriptTypeBuiltin,
		ScriptCode:     "Builtin_NetworkPolicy_018",
		TimeoutSeconds: 60, // 需要检查Pod选择器匹配
		Script: `
			local nps, err = kubectl:GVK("networking.k8s.io", "v1", "NetworkPolicy"):AllNamespace(""):List()
			if err then print("获取 NetworkPolicy 失败: " .. tostring(err)) return end
			for _, np in ipairs(nps) do
				if np.spec and np.spec.podSelector and (not np.spec.podSelector.matchLabels or next(np.spec.podSelector.matchLabels) == nil) then
					check_event("失败", "NetworkPolicy '" .. np.metadata.name .. "' 允许所有 Pod", {namespace=np.metadata.namespace, name=np.metadata.name})
				else
					local selector = ""
					if np.spec and np.spec.podSelector and np.spec.podSelector.matchLabels then
						for k, v in pairs(np.spec.podSelector.matchLabels) do
							if selector ~= "" then selector = selector .. "," end
							selector = selector .. k .. "=" .. v
						end
					end
					if selector ~= "" then
						local pods, err = kubectl:GVK("", "v1", "Pod"):Namespace(np.metadata.namespace):WithLabelSelector(selector):List()
						if not err and pods and #pods.items == 0 then
							check_event("失败", "NetworkPolicy '" .. np.metadata.name .. "' 未作用于任何 Pod", {namespace=np.metadata.namespace, name=np.metadata.name})
						end
					end
				end
			end
			print("NetworkPolicy 合规性检查完成")
		`,
	},
	{
		Name:           "Node 合规性检查",
		Description:    "检查 Node 的 Condition 状态，非 Ready/EtcdIsVoter 且状态异常时报警。",
		Group:          "",
		Version:        "v1",
		Kind:           "Node",
		ScriptType:     constants.LuaScriptTypeBuiltin,
		ScriptCode:     "Builtin_Node_019",
		TimeoutSeconds: 45, // Node状态检查相对简单
		Script: `
			local nodes, err = kubectl:GVK("", "v1", "Node"):AllNamespace(""):List()
			if err then print("获取 Node 失败: " .. tostring(err)) return end
			for _, node in ipairs(nodes) do
				if node.status and node.status.conditions then
					for _, cond in ipairs(node.status.conditions) do
						if cond.type == "Ready" then
							if cond.status ~= "True" then
								check_event("失败", node.metadata.name .. " Ready 状态异常: " .. (cond.reason or "") .. " - " .. (cond.message or ""), {name=node.metadata.name, type=cond.type, reason=cond.reason, message=cond.message})
							end
						elseif cond.type == "EtcdIsVoter" then
							-- 跳过 k3s 特有的 EtcdIsVoter
						else
							if cond.status ~= "False" then
								check_event("失败", node.metadata.name .. " " .. cond.type .. " 状态异常: " .. (cond.reason or "") .. " - " .. (cond.message or ""), {name=node.metadata.name, type=cond.type, reason=cond.reason, message=cond.message})
							end
						end
					end
				end
			end
			print("Node 合规性检查完成")
		`,
	},
	{
		Name:           "Pod 合规性检查",
		Description:    "检查 Pod 的 Pending、调度失败、CrashLoopBackOff、终止异常、ReadinessProbe 失败等状态。",
		Group:          "",
		Version:        "v1",
		Kind:           "Pod",
		ScriptType:     constants.LuaScriptTypeBuiltin,
		ScriptCode:     "Builtin_Pod_020",
		TimeoutSeconds: 120, // Pod状态检查复杂，需要检查多种状态
		Script: `
			local pods, err = kubectl:GVK("", "v1", "Pod"):AllNamespace(""):List()
			if err then print("获取 Pod 失败: " .. tostring(err)) return end
			for _, pod in ipairs(pods) do
				if pod.status and pod.status.phase == "Pending" and pod.status.conditions then
					for _, cond in ipairs(pod.status.conditions) do
						if cond.type == "PodScheduled" and cond.reason == "Unschedulable" and cond.message and cond.message ~= "" then
							check_event("失败", cond.message, {namespace=pod.metadata.namespace, name=pod.metadata.name})
						end
					end
				end
				local function check_container_statuses(statuses, phase)
					if not statuses then return end
					for _, cs in ipairs(statuses) do
						if cs.state and cs.state.waiting then
							if cs.state.waiting.reason == "CrashLoopBackOff" and cs.lastState and cs.lastState.terminated then
								check_event("失败", "CrashLoopBackOff: 上次终止原因 " .. (cs.lastState.terminated.reason or "") .. " 容器=" .. cs.name .. " pod=" .. pod.metadata.name, {namespace=pod.metadata.namespace, name=pod.metadata.name, container=cs.name})
							elseif cs.state.waiting.reason and (cs.state.waiting.reason == "ImagePullBackOff" or cs.state.waiting.reason == "ErrImagePull" or cs.state.waiting.reason == "CreateContainerConfigError" or cs.state.waiting.reason == "CreateContainerError" or cs.state.waiting.reason == "RunContainerError" or cs.state.waiting.reason == "InvalidImageName") then
								check_event("失败", cs.state.waiting.message or (cs.state.waiting.reason .. " 容器=" .. cs.name .. " pod=" .. pod.metadata.name), {namespace=pod.metadata.namespace, name=pod.metadata.name, container=cs.name})
							end
						elseif cs.state and cs.state.terminated and cs.state.terminated.exitCode and cs.state.terminated.exitCode ~= 0 then
							check_event("失败", "终止异常: " .. (cs.state.terminated.reason or "Unknown") .. " exitCode=" .. tostring(cs.state.terminated.exitCode) .. " 容器=" .. cs.name .. " pod=" .. pod.metadata.name, {namespace=pod.metadata.namespace, name=pod.metadata.name, container=cs.name, exitCode=cs.state.terminated.exitCode})
						elseif cs.ready == false and phase == "Running" then
							check_event("失败", "容器未就绪: " .. cs.name .. " pod=" .. pod.metadata.name, {namespace=pod.metadata.namespace, name=pod.metadata.name, container=cs.name})
						end
					end
				end
				if pod.status then
					check_container_statuses(pod.status.initContainerStatuses, pod.status.phase)
					check_container_statuses(pod.status.containerStatuses, pod.status.phase)
				end
			end
			print("Pod 合规性检查完成")
		`,
	},
	{
		Name:           "PVC 合规性检查",
		Description:    "检查 PVC Pending 状态下的 ProvisioningFailed 事件。",
		Group:          "",
		Version:        "v1",
		Kind:           "PersistentVolumeClaim",
		ScriptType:     constants.LuaScriptTypeBuiltin,
		ScriptCode:     "Builtin_PVC_021",
		TimeoutSeconds: 60, // 需要检查Event事件
		Script: `
			local pvcs, err = kubectl:GVK("", "v1", "PersistentVolumeClaim"):AllNamespace(""):List()
			if err then print("获取 PVC 失败: " .. tostring(err)) return end
			for _, pvc in ipairs(pvcs) do
				if pvc.status and pvc.status.phase == "Pending" then
					local events, err = kubectl:GVK("", "v1", "Event"):Namespace(pvc.metadata.namespace):WithFieldSelector("involvedObject.name=" .. pvc.metadata.name):List()
					if not err and events and events.items then
						for _, evt in ipairs(events.items) do
							if evt.reason == "ProvisioningFailed" and evt.message and evt.message ~= "" then
								check_event("失败", evt.message, {namespace=pvc.metadata.namespace, name=pvc.metadata.name})
							end
						end
					end
				end
			end
			print("PVC 合规性检查完成")
		`,
	},
	{
		Name:           "ReplicaSet 合规性检查",
		Description:    "检测副本数为0且有 FailedCreate 的 ReplicaFailure。",
		Group:          "apps",
		Version:        "v1",
		Kind:           "ReplicaSet",
		ScriptType:     constants.LuaScriptTypeBuiltin,
		ScriptCode:     "Builtin_ReplicaSet_022",
		TimeoutSeconds: 45, // ReplicaSet状态检查相对简单
		Script: `
			local rss, err = kubectl:GVK("apps", "v1", "ReplicaSet"):AllNamespace(""):List()
			if err then print("获取 ReplicaSet 失败: " .. tostring(err)) return end
			for _, rs in ipairs(rss) do
				if rs.status and rs.status.replicas == 0 and rs.status.conditions then
					for _, cond in ipairs(rs.status.conditions) do
						if cond.type == "ReplicaFailure" and cond.reason == "FailedCreate" then
							check_event("失败", cond.message or "ReplicaSet 副本创建失败", {namespace=rs.metadata.namespace, name=rs.metadata.name})
						end
					end
				end
			end
			print("ReplicaSet 合规性检查完成")
		`,
	},
	{
		Name:           "Security ServiceAccount 默认账户使用检测",
		Description:    "检测 default ServiceAccount 是否被 Pod 使用。",
		Group:          "core",
		Version:        "v1",
		Kind:           "ServiceAccount",
		ScriptType:     constants.LuaScriptTypeBuiltin,
		ScriptCode:     "Builtin_Security_SA_023",
		TimeoutSeconds: 60, // 需要检查Pod使用情况
		Script: `
			local sas, err = kubectl:GVK("", "v1", "ServiceAccount"):AllNamespace(""):List()
			if err then print("获取 ServiceAccount 失败: " .. tostring(err)) return end
			for _, sa in ipairs(sas) do
				if sa.metadata and sa.metadata.name == "default" then
					local pods, err = kubectl:GVK("", "v1", "Pod"):Namespace(sa.metadata.namespace):List()
					if not err and pods then
						local defaultSAUsers = {}
						for _, pod in ipairs(pods) do
							if pod.spec and pod.spec.serviceAccountName == "default" then
								table.insert(defaultSAUsers, pod.metadata.name)
							end
						end
						if #defaultSAUsers > 0 then
							check_event("失败", "Default service account 被以下 Pod 使用: " .. table.concat(defaultSAUsers, ", "), {namespace=sa.metadata.namespace, name=sa.metadata.name})
						end
					end
				end
			end
			print("Security ServiceAccount 检查完成")
		`,
	},
	{
		Name:           "Security RoleBinding 通配符检测",
		Description:    "检测 RoleBinding 关联的 Role 是否包含通配符权限。",
		Group:          "rbac.authorization.k8s.io",
		Version:        "v1",
		Kind:           "RoleBinding",
		ScriptType:     constants.LuaScriptTypeBuiltin,
		ScriptCode:     "Builtin_Security_RoleBinding_024",
		TimeoutSeconds: 75, // 需要检查Role权限规则
		Script: `
			local rbs, err = kubectl:GVK("rbac.authorization.k8s.io", "v1", "RoleBinding"):AllNamespace(""):List()
			if err then print("获取 RoleBinding 失败: " .. tostring(err)) return end
			for _, rb in ipairs(rbs) do
				if rb.roleRef and rb.roleRef.kind == "Role" and rb.roleRef.name then
					local role, err = kubectl:GVK("rbac.authorization.k8s.io", "v1", "Role"):Namespace(rb.metadata.namespace):Name(rb.roleRef.name):Get()
					if not err and role and role.rules then
						for _, rule in ipairs(role.rules) do
							local function containsWildcard(arr)
								if not arr then return false end
								for _, v in ipairs(arr) do if v == "*" then return true end end
								return false
							end
							if containsWildcard(rule.verbs) or containsWildcard(rule.resources) then
								check_event("失败", "RoleBinding '" .. rb.metadata.name .. "' 关联的 Role '" .. role.metadata.name .. "' 存在通配符权限", {namespace=rb.metadata.namespace, name=rb.metadata.name, role=role.metadata.name})
							end
						end
					end
				end
			end
			print("Security RoleBinding 检查完成")
		`,
	},
	{
		Name:           "Security Pod 安全上下文检测",
		Description:    "检测 Pod 是否存在特权容器或缺少安全上下文。",
		Group:          "core",
		Version:        "v1",
		Kind:           "Pod",
		ScriptType:     constants.LuaScriptTypeBuiltin,
		ScriptCode:     "Builtin_Security_Pod_025",
		TimeoutSeconds: 90, // 需要检查所有Pod的安全上下文
		Script: `
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
		`,
	},
	{
		Name:           "StatefulSet 合规性检查",
		Description:    "检测 StatefulSet 关联的 Service、StorageClass 是否存在及 Pod 状态。",
		Group:          "apps",
		Version:        "v1",
		Kind:           "StatefulSet",
		ScriptType:     constants.LuaScriptTypeBuiltin,
		ScriptCode:     "Builtin_StatefulSet_026",
		TimeoutSeconds: 120, // 需要检查Service、StorageClass和Pod状态，较为复杂
		Script: `
			local stss, err = kubectl:GVK("apps", "v1", "StatefulSet"):AllNamespace(""):List()
			if err then print("获取 StatefulSet 失败: " .. tostring(err)) return end
			for _, sts in ipairs(stss) do
				if sts.spec and sts.spec.serviceName then
					local svc, err = kubectl:GVK("", "v1", "Service"):Namespace(sts.metadata.namespace):Name(sts.spec.serviceName):Get()
					if err or not svc then
						check_event("失败", "StatefulSet 使用的 Service '" .. sts.metadata.namespace .. "/" .. sts.spec.serviceName .. "' 不存在", {namespace=sts.metadata.namespace, name=sts.metadata.name, service=sts.spec.serviceName})
					end
				end
				if sts.spec and sts.spec.volumeClaimTemplates then
					for _, vct in ipairs(sts.spec.volumeClaimTemplates) do
						if vct.spec and vct.spec.storageClassName then
							local sc, err = kubectl:GVK("storage.k8s.io", "v1", "StorageClass"):Name(vct.spec.storageClassName):Get()
							if err or not sc then
								check_event("失败", "StatefulSet 使用的 StorageClass '" .. vct.spec.storageClassName .. "' 不存在", {namespace=sts.metadata.namespace, name=sts.metadata.name, storageClass=vct.spec.storageClassName})
							end
						end
					end
				end
				if sts.spec and sts.spec.replicas and sts.status and sts.status.availableReplicas and sts.spec.replicas ~= sts.status.availableReplicas then
					for i = 0, sts.spec.replicas - 1 do
						local podName = sts.metadata.name .. "-" .. tostring(i)
						local pod, err = kubectl:GVK("", "v1", "Pod"):Namespace(sts.metadata.namespace):Name(podName):Get()
						if err or not pod then
							if i == 0 then
								local events, err = kubectl:GVK("", "v1", "Event"):Namespace(sts.metadata.namespace):WithFieldSelector("involvedObject.name=" .. sts.metadata.name):List()
								if not err and events and events.items then
									for _, evt in ipairs(events.items) do
										if evt.type ~= "Normal" and evt.message and evt.message ~= "" then
											check_event("失败", evt.message, {namespace=sts.metadata.namespace, name=sts.metadata.name})
										end
									end
								end
							end
							break
						end
						if pod.status and pod.status.phase ~= "Running" then
							check_event("失败", "StatefulSet 的 Pod '" .. pod.metadata.name .. "' 不在 Running 状态", {namespace=sts.metadata.namespace, name=sts.metadata.name, pod=pod.metadata.name, phase=pod.status.phase})
							break
						end
					end
				end
			end
			print("StatefulSet 合规性检查完成")
		`,
	},
	{
		Name:           "StorageClass 合规性检查",
		Description:    "检测 StorageClass 是否使用了已废弃的 provisioner，及是否存在多个默认 StorageClass。",
		Group:          "storage.k8s.io",
		Version:        "v1",
		Kind:           "StorageClass",
		ScriptType:     constants.LuaScriptTypeBuiltin,
		ScriptCode:     "Builtin_StorageClass_027",
		TimeoutSeconds: 30, // StorageClass检查相对简单
		Script: `
			local scs, err = kubectl:GVK("storage.k8s.io", "v1", "StorageClass"):AllNamespace(""):List()
			if err then print("获取 StorageClass 失败: " .. tostring(err)) return end
			local defaultCount = 0
			for _, sc in ipairs(scs) do
				if sc.provisioner == "kubernetes.io/no-provisioner" then
					check_event("失败", "StorageClass '" .. sc.metadata.name .. "' 使用了已废弃的 provisioner 'kubernetes.io/no-provisioner'", {name=sc.metadata.name})
				end
				if sc.metadata.annotations and sc.metadata.annotations["storageclass.kubernetes.io/is-default-class"] == "true" then
					defaultCount = defaultCount + 1
				end
			end
			if defaultCount > 1 then
				check_event("失败", "存在多个默认 StorageClass (" .. tostring(defaultCount) .. ")，可能导致混淆", {})
			end
			print("StorageClass 合规性检查完成")
		`,
	},
	{
		Name:           "PersistentVolume 合规性检查",
		Description:    "检测 PV 是否为 Released/Failed 状态，及容量小于 1Gi。",
		Group:          "core",
		Version:        "v1",
		Kind:           "PersistentVolume",
		ScriptType:     constants.LuaScriptTypeBuiltin,
		ScriptCode:     "Builtin_PV_028",
		TimeoutSeconds: 45, // PV状态和容量检查
		Script: `
			local pvs, err = kubectl:GVK("", "v1", "PersistentVolume"):AllNamespace(""):List()
			if err then print("获取 PersistentVolume 失败: " .. tostring(err)) return end
			for _, pv in ipairs(pvs) do
				if pv.status and pv.status.phase == "Released" then
					check_event("失败", "PersistentVolume '" .. pv.metadata.name .. "' 处于 Released 状态，应及时清理", {name=pv.metadata.name})
				end
				if pv.status and pv.status.phase == "Failed" then
					check_event("失败", "PersistentVolume '" .. pv.metadata.name .. "' 处于 Failed 状态", {name=pv.metadata.name})
				end
				if pv.spec and pv.spec.capacity and pv.spec.capacity.storage then
					local function parseGi(val)
						local n = tonumber(val:match("%d+"))
						if val:find("Gi") then return n end
						if val:find("Mi") then return n and n/1024 or 0 end
						return 0
					end
					if parseGi(pv.spec.capacity.storage) < 1 then
						check_event("失败", "PersistentVolume '" .. pv.metadata.name .. "' 容量过小 (" .. pv.spec.capacity.storage .. ")", {name=pv.metadata.name, capacity=pv.spec.capacity.storage})
					end
				end
			end
			print("PersistentVolume 合规性检查完成")
		`,
	},
	{
		Name:        "PersistentVolumeClaim 合规性检查",
		Description: "检测 PVC Pending/Lost 状态、容量小于 1Gi、无 StorageClass。",
		Group:       "core",
		Version:     "v1",
		Kind:        "PersistentVolumeClaim",
		ScriptType:  constants.LuaScriptTypeBuiltin,
		ScriptCode:  "Builtin_PVC_029",
		Script: `
			local pvcs, err = kubectl:GVK("", "v1", "PersistentVolumeClaim"):AllNamespace(""):List()
			if err then print("获取 PVC 失败: " .. tostring(err)) return end
			for _, pvc in ipairs(pvcs) do
				if pvc.status and pvc.status.phase == "Pending" then
					check_event("失败", "PersistentVolumeClaim '" .. pvc.metadata.name .. "' 处于 Pending 状态", {namespace=pvc.metadata.namespace, name=pvc.metadata.name})
				elseif pvc.status and pvc.status.phase == "Lost" then
					check_event("失败", "PersistentVolumeClaim '" .. pvc.metadata.name .. "' 处于 Lost 状态", {namespace=pvc.metadata.namespace, name=pvc.metadata.name})
				else
					if pvc.spec and pvc.spec.resources and pvc.spec.resources.requests and pvc.spec.resources.requests.storage then
						local function parseGi(val)
							local n = tonumber(val:match("%d+"))
							if val:find("Gi") then return n end
							if val:find("Mi") then return n and n/1024 or 0 end
							return 0
						end
						if parseGi(pvc.spec.resources.requests.storage) < 1 then
							check_event("失败", "PersistentVolumeClaim '" .. pvc.metadata.name .. "' 容量过小 (" .. pvc.spec.resources.requests.storage .. ")", {namespace=pvc.metadata.namespace, name=pvc.metadata.name, capacity=pvc.spec.resources.requests.storage})
						end
					end
					if (not pvc.spec or not pvc.spec.storageClassName) and (not pvc.spec or not pvc.spec.volumeName or pvc.spec.volumeName == "") then
						check_event("失败", "PersistentVolumeClaim '" .. pvc.metadata.name .. "' 未指定 StorageClass", {namespace=pvc.metadata.namespace, name=pvc.metadata.name})
					end
				end
			end
			print("PersistentVolumeClaim 合规性检查完成")
		`,
	},
	{
		Name:        "ValidatingWebhookConfiguration 合规性检查",
		Description: "检查 ValidatingWebhookConfiguration 的 webhook 指向的 Service 是否存在、是否有活跃 Pod、Pod 状态。",
		Group:       "admissionregistration.k8s.io",
		Version:     "v1",
		Kind:        "ValidatingWebhookConfiguration",
		ScriptType:  constants.LuaScriptTypeBuiltin,
		ScriptCode:  "Builtin_ValidatingWebhook_030",
		Script: `
			local vwcs, err = kubectl:GVK("admissionregistration.k8s.io", "v1", "ValidatingWebhookConfiguration"):AllNamespace(""):List()
			if err then print("获取 ValidatingWebhookConfiguration 失败: " .. tostring(err)) return end
			for _, vwc in ipairs(vwcs) do
				if vwc.webhooks then
					for _, webhook in ipairs(vwc.webhooks) do
						if webhook.clientConfig and webhook.clientConfig.service then
							local svc = webhook.clientConfig.service
							local service, err = kubectl:GVK("", "v1", "Service"):Namespace(svc.namespace):Name(svc.name):Get()
							if err or not service then
								check_event("失败", "ValidatingWebhook " .. webhook.name .. " 指向的 Service '" .. svc.namespace .. "/" .. svc.name .. "' 不存在", {namespace=svc.namespace, name=svc.name, webhook=webhook.name})
							else
								if service.spec and service.spec.selector and next(service.spec.selector) ~= nil then
									local selector = ""
									for k, v in pairs(service.spec.selector) do
										if selector ~= "" then selector = selector .. "," end
										selector = selector .. k .. "=" .. v
									end
									local pods, err = kubectl:GVK("", "v1", "Pod"):Namespace(svc.namespace):WithLabelSelector(selector):List()
									if not err and pods and #pods.items == 0 then
										check_event("失败", "ValidatingWebhook " .. webhook.name .. " 指向的 Service '" .. svc.namespace .. "/" .. svc.name .. "' 没有活跃 Pod", {namespace=svc.namespace, name=svc.name, webhook=webhook.name})
									end
									if pods and pods.items then
										for _, pod in ipairs(pods.items) do
											if pod.status and pod.status.phase ~= "Running" then
												check_event("失败", "ValidatingWebhook " .. webhook.name .. " 指向的 Pod '" .. pod.metadata.name .. "' 状态为 " .. (pod.status.phase or "未知") , {namespace=svc.namespace, name=svc.name, webhook=webhook.name, pod=pod.metadata.name, phase=pod.status.phase})
											end
										end
									end
								end
							end
						end
					end
				end
			end
			print("ValidatingWebhookConfiguration 合规性检查完成")
		`,
	},
	{
		Name:        "Pod 日志错误检测",
		Description: "检查某一个 Pod 的最近日志是否包含指定关键字，若包含则认为检测失败。",
		Group:       "",
		Version:     "v1",
		Kind:        "Pod",
		ScriptType:  constants.LuaScriptTypeBuiltin,
		ScriptCode:  "Builtin_Pod_Log_Error_031",
		Script: `
			-- 示例：根据已知 Deployment 名称与命名空间，按其 selector 获取 Pod 列表并检查日志
			-- 请按需修改以下四个变量
			local deployName = "your-deploy-name"
			local namespace = "default"
			local keyword = "ERROR"  -- 默认关键字为 "ERROR"，可按需改为要检测的关键字
			local tailLines = 200    -- 默认读取最近 200 行日志，按需调整

			-- 获取 Deployment 对象
			local dep, derr = kubectl:GVK("apps", "v1", "Deployment"):Namespace(namespace):Name(deployName):Get()
			if derr ~= nil or not dep then
				print("获取 Deployment 失败: " .. tostring(derr))
				return
			end

			-- 从 Deployment 的 selector.matchLabels 构建 LabelSelector
			local matchLabels = dep.spec and dep.spec.selector and dep.spec.selector.matchLabels or nil
			if not matchLabels then
				print("Deployment 未定义 selector.matchLabels，无法按标签筛选 Pod")
				return
			end
			local labelSelector = ""
			for k, v in pairs(matchLabels) do
				if labelSelector ~= "" then labelSelector = labelSelector .. "," end
				labelSelector = labelSelector .. k .. "=" .. v
			end

			-- 按 Deployment 的 selector 在同一命名空间获取 Pod 列表
			local pods, perr = kubectl:GVK("", "v1", "Pod"):Namespace(namespace):Cache(10):WithLabelSelector(labelSelector):List()
			if perr ~= nil then
				print("获取 Pod 列表失败: " .. tostring(perr))
				return
			end
			if not pods or #pods == 0 then
				print("未找到与 Deployment 匹配的 Pod: " .. namespace .. "/" .. deployName .. ", selector=" .. labelSelector)
				return
			end

			local foundError = false
			for _, pod in ipairs(pods) do
				local ns = pod.metadata.namespace
				local name = pod.metadata.name
				local containerName = nil
				if pod.spec and pod.spec.containers and #pod.spec.containers > 0 then
					containerName = pod.spec.containers[1].name
				end
				local opts = { tailLines = tailLines }
				if containerName ~= nil then
					opts.container = containerName
				end
				local logs, lerr = kubectl:GVK("", "v1", "Pod"):Namespace(ns):Name(name):GetLogs(opts)
				if lerr ~= nil then
					print("获取 Pod 日志失败: " .. tostring(lerr))
				else
					local logStr = (type(logs) == "string") and logs or tostring(logs)
					if logStr and string.find(logStr, keyword) ~= nil then
						foundError = true
						check_event("失败", "Pod 日志包含关键字 '" .. keyword .. "'", {namespace=ns, name=name, container=containerName, keyword=keyword})
					else
						print("Pod " .. ns .. "/" .. name .. " 最近日志未发现 '" .. keyword .. "'")
					end
				end
			end

			if not foundError then
				print("Deployment '" .. namespace .. "/" .. deployName .. "' 关联 Pod 的日志检查完成，未发现 '" .. keyword .. "'")
			end
			print("Pod 日志错误检测完成")
		`,
	},
	{
		Name:           "Pod 资源用量检查",
		Description:    "检查指定 Pod 的资源用量情况，包括 CPU 和内存的请求、限制、实时用量等信息",
		Group:          "",
		Version:        "v1",
		Kind:           "Pod",
		ScriptType:     constants.LuaScriptTypeBuiltin,
		ScriptCode:     "Builtin_Pod_ResourceUsage_032",
		TimeoutSeconds: 90, // 需要获取Pod资源用量数据，包含复杂的计算逻辑
		Script: `
			-- =============================
-- 🧩 Pod 资源用量检查脚本（JSON格式输出 + 比例修正）
-- =============================

-- 请修改以下变量为您要检查的 Pod 信息
local podName = "k8m-c6dccfb-qm7cp"  -- 要检查的 Pod 名称
local podNamespace = "k8m"           -- Pod 所在的命名空间

-- =============================
-- 可配置的告警阈值
-- =============================
-- 
-- 配置说明：
-- - cpuThreshold: CPU 使用率告警阈值，取值范围 0.0-1.0（例如：0.8 表示 80%）
-- - memoryThreshold: 内存使用率告警阈值，取值范围 0.0-1.0（例如：0.9 表示 90%）
-- 
-- 建议值：
-- - 生产环境：CPU 0.7-0.8，内存 0.8-0.9
-- - 测试环境：CPU 0.8-0.9，内存 0.9-0.95
-- - 开发环境：可适当放宽至 CPU 0.9，内存 0.95
local cpuThreshold = 0.8    -- CPU 使用率告警阈值（80%）
local memoryThreshold = 0.9 -- 内存使用率告警阈值（90%）

-- =============================
-- 工具函数
-- =============================

-- 将 Lua table 转为美化 JSON 字符串
local function to_json(tbl, indent)
    indent = indent or 0
    local padding = string.rep("  ", indent)
    if type(tbl) ~= "table" then
        if type(tbl) == "string" then
            return string.format("%q", tbl)
        else
            return tostring(tbl)
        end
    end
    local lines = {"{"}
    for k, v in pairs(tbl) do
        local key = string.format("%q", tostring(k))
        local val = to_json(v, indent + 1)
        local comma = (next(tbl, k) ~= nil) and "," or ""
        table.insert(lines, string.rep("  ", indent + 1) .. key .. ": " .. val .. comma)
    end
    table.insert(lines, padding .. "}")
    return table.concat(lines, "\n")
end

-- 字节换算为人类可读单位
local function human_bytes(n)
    if type(n) ~= "number" then return tostring(n) end
    local units = {"B", "KiB", "MiB", "GiB", "TiB"}
    local i = 1
    while n >= 1024 and i < #units do
        n = n / 1024
        i = i + 1
    end
    return string.format("%.2f %s", n, units[i])
end

-- 获取 allocatable.memory 的值
local function get_allocatable_memory(r)
    if not r then return nil end
    if r.memory and r.memory.allocatable then
        return tonumber(r.memory.allocatable)
    end
    if r.allocatable and r.allocatable.memory then
        return tonumber(r.allocatable.memory)
    end
    return nil
end

-- =============================
-- 获取 Pod 资源用量
-- =============================

local resourceUsage, err = kubectl:GVK("", "v1", "Pod"):Namespace(podNamespace):Name(podName):GetPodResourceUsage()
if err then
    print("获取 Pod 资源用量失败: " .. tostring(err))
    return
end

if not resourceUsage then
    print("Pod " .. podNamespace .. "/" .. podName .. " 资源用量信息为空")
    return
end

print("=== Pod 资源用量原始数据（JSON 格式） ===")
print(to_json(resourceUsage))

print("\n=== Pod 资源用量检查结果 ===")
print("Pod: " .. podNamespace .. "/" .. podName)

-- =============================
-- CPU 检查
-- =============================
if resourceUsage.cpu then
    print("\n--- CPU 资源 ---")
    if resourceUsage.cpu.requests then
        print("CPU 请求量: " .. tostring(resourceUsage.cpu.requests))
    end
    if resourceUsage.cpu.limits then
        print("CPU 限制量: " .. tostring(resourceUsage.cpu.limits))
    end
    if resourceUsage.cpu.realtime then
        print("CPU 实时用量: " .. tostring(resourceUsage.cpu.realtime))
    elseif resourceUsage.realtime and resourceUsage.realtime.cpu then
        print("CPU 实时用量: " .. tostring(resourceUsage.realtime.cpu))
    end
    if resourceUsage.cpu.allocatable then
        print("CPU 可分配量: " .. tostring(resourceUsage.cpu.allocatable))
    elseif resourceUsage.allocatable and resourceUsage.allocatable.cpu then
        print("CPU 可分配量: " .. tostring(resourceUsage.allocatable.cpu))
    end

    local cpuUsage = nil
    if resourceUsage.cpu.usageFractions then
        cpuUsage = tonumber(resourceUsage.cpu.usageFractions)
    elseif resourceUsage.usageFractions and resourceUsage.usageFractions.cpu and resourceUsage.usageFractions.cpu.realtimeFraction then
        cpuUsage = tonumber(resourceUsage.usageFractions.cpu.realtimeFraction)
    end

    if cpuUsage then
        if cpuUsage > 1 then
            print(string.format("CPU 使用率 (原始): %.2f%%", cpuUsage))
            cpuUsage = cpuUsage / 100
        end
        print(string.format("CPU 使用率: %.2f%%", cpuUsage * 100))
        if cpuUsage > cpuThreshold then
            check_event("警告", "Pod " .. podNamespace .. "/" .. podName .. " CPU 使用率过高: " .. string.format("%.2f%%", cpuUsage * 100), {namespace=podNamespace, name=podName, cpuUsage=cpuUsage})
        end
    end
end

-- =============================
-- 内存检查（修正版）
-- =============================
if resourceUsage.memory or resourceUsage.allocatable then
    print("\n--- 内存资源 ---")

    local memRealtime = nil
    if resourceUsage.memory and resourceUsage.memory.realtime then
        memRealtime = tonumber(resourceUsage.memory.realtime)
    elseif resourceUsage.realtime and resourceUsage.realtime.memory then
        memRealtime = tonumber(resourceUsage.realtime.memory)
    end

    local memAllocatable = get_allocatable_memory(resourceUsage)
    local memRequests = resourceUsage.memory and resourceUsage.memory.requests or nil
    local memLimits = resourceUsage.memory and resourceUsage.memory.limits or nil

    print("内存请求量: " .. tostring(memRequests or "(未设置)"))
    print("内存限制量: " .. tostring(memLimits or "(未设置)"))

    if memRealtime then
        print("内存实时用量: " .. human_bytes(memRealtime))
    else
        print("内存实时用量: (无数据)")
    end

    if memAllocatable then
        print("内存可分配量: " .. human_bytes(memAllocatable))
    else
        print("内存可分配量: (无数据)")
    end

    -- 重新计算 fraction
    local recomputedFraction = nil
    if memRealtime and memAllocatable and memAllocatable > 0 then
        recomputedFraction = memRealtime / memAllocatable
    end

    if recomputedFraction then
        print(string.format("内存使用率: %.2f%%", recomputedFraction * 100))
        if recomputedFraction > memoryThreshold then
            check_event("警告", "Pod " .. podNamespace .. "/" .. podName .. " 内存使用率过高: " .. string.format("%.2f%%", recomputedFraction * 100), {namespace=podNamespace, name=podName, memoryUsage=recomputedFraction})
        end
    else
        local rawUF = nil
        if resourceUsage.usageFractions and resourceUsage.usageFractions.memory and resourceUsage.usageFractions.memory.realtimeFraction then
            rawUF = tonumber(resourceUsage.usageFractions.memory.realtimeFraction)
        end
        if rawUF then
            if rawUF > 1 then
                print(string.format("内存使用率 (来源 usageFractions): %.2f%% (已推测为百分比)", rawUF))
                rawUF = rawUF / 100
            else
                print(string.format("内存使用率: %.2f%%", rawUF * 100))
            end
            if rawUF > memoryThreshold then
                check_event("警告", "Pod " .. podNamespace .. "/" .. podName .. " 内存使用率过高: " .. string.format("%.2f%%", rawUF * 100), {namespace=podNamespace, name=podName, memoryUsage=rawUF})
            end
        else
            print("内存使用率: (无法计算 —— 缺少数据)")
        end
    end
end

-- =============================
-- 检查 requests / limits 配置
-- =============================
local hasRequests = (resourceUsage.cpu and resourceUsage.cpu.requests) or (resourceUsage.memory and resourceUsage.memory.requests)
local hasLimits = (resourceUsage.cpu and resourceUsage.cpu.limits) or (resourceUsage.memory and resourceUsage.memory.limits)

if not hasRequests then
    check_event("失败", "Pod " .. podNamespace .. "/" .. podName .. " 未配置资源请求量 (requests)", {namespace=podNamespace, name=podName})
end

if not hasLimits then
    check_event("失败", "Pod " .. podNamespace .. "/" .. podName .. " 未配置资源限制量 (limits)", {namespace=podNamespace, name=podName})
end

print("\n✅ Pod 资源用量检查完成")

		`,
	},
}
