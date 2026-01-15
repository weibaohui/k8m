package models

import (
	"github.com/weibaohui/k8m/pkg/constants"
	"k8s.io/klog/v2"
)

var builtinLuaScriptsPodProbeCompliance = []InspectionLuaScript{
	{
		Name:           "Pod 探针配置合规性检查 | 未配置存活/就绪探针",
		Description:    "检查容器是否未配置 LivenessProbe/ReadinessProbe，避免 Pod 误杀或服务未就绪即接收流量。",
		Group:          "",
		Version:        "v1",
		Kind:           "Pod",
		ScriptType:     constants.LuaScriptTypeBuiltin,
		ScriptCode:     "Builtin_Probe_001",
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
		`,
	},
	{
		Name:           "Pod 探针配置合规性检查 | 参数不合理",
		Description:    "检查 Liveness/Readiness 探针参数是否不合理（如 timeoutSeconds 过短、periodSeconds 过长）。",
		Group:          "",
		Version:        "v1",
		Kind:           "Pod",
		ScriptType:     constants.LuaScriptTypeBuiltin,
		ScriptCode:     "Builtin_Probe_002",
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

			local function has_probe_action(probe)
				if not probe then
					return false
				end
				if probe.httpGet then return true end
				if probe.exec then return true end
				if probe.tcpSocket then return true end
				if probe.grpc then return true end
				return false
			end

			local function check_probe_params(ns, podName, containerName, probe, probeType)
				if not probe then
					return
				end

				if not has_probe_action(probe) then
					check_event(
						"失败",
						"Pod " .. ns .. "/" .. podName .. " 容器 " .. containerName .. " 的 " .. probeType .. " 探针未配置检查方式（httpGet/exec/tcpSocket/grpc）",
						{ namespace = ns, name = podName, container = containerName, probe = probeType }
					)
					return
				end

				local timeout = probe.timeoutSeconds
				if timeout ~= nil and tonumber(timeout) ~= nil and tonumber(timeout) < 2 then
					check_event(
						"失败",
						"Pod " .. ns .. "/" .. podName .. " 容器 " .. containerName .. " 的 " .. probeType .. " 探针 timeoutSeconds 过短(" .. tostring(timeout) .. "s)，可能导致误杀/误判",
						{ namespace = ns, name = podName, container = containerName, probe = probeType, field = "timeoutSeconds", value = timeout }
					)
				end

				local period = probe.periodSeconds
				if period ~= nil and tonumber(period) ~= nil and tonumber(period) > 60 then
					check_event(
						"失败",
						"Pod " .. ns .. "/" .. podName .. " 容器 " .. containerName .. " 的 " .. probeType .. " 探针 periodSeconds 过长(" .. tostring(period) .. "s)，可能导致故障发现过慢或就绪切换不及时",
						{ namespace = ns, name = podName, container = containerName, probe = probeType, field = "periodSeconds", value = period }
					)
				end

				local initialDelay = probe.initialDelaySeconds
				if probeType == "liveness" and initialDelay ~= nil and tonumber(initialDelay) ~= nil and tonumber(initialDelay) < 5 then
					check_event(
						"失败",
						"Pod " .. ns .. "/" .. podName .. " 容器 " .. containerName .. " 的 liveness 探针 initialDelaySeconds 过短(" .. tostring(initialDelay) .. "s)，可能导致启动阶段误杀",
						{ namespace = ns, name = podName, container = containerName, probe = probeType, field = "initialDelaySeconds", value = initialDelay }
					)
				end

				local failure = probe.failureThreshold
				if failure ~= nil and tonumber(failure) ~= nil then
					if probeType == "liveness" and tonumber(failure) < 3 then
						check_event(
							"失败",
							"Pod " .. ns .. "/" .. podName .. " 容器 " .. containerName .. " 的 liveness 探针 failureThreshold 过低(" .. tostring(failure) .. ")，容易误杀",
							{ namespace = ns, name = podName, container = containerName, probe = probeType, field = "failureThreshold", value = failure }
						)
					end
					if probeType == "readiness" and tonumber(failure) < 2 then
						check_event(
							"失败",
							"Pod " .. ns .. "/" .. podName .. " 容器 " .. containerName .. " 的 readiness 探针 failureThreshold 过低(" .. tostring(failure) .. ")，容易频繁就绪抖动",
							{ namespace = ns, name = podName, container = containerName, probe = probeType, field = "failureThreshold", value = failure }
						)
					end
				end

				local success = probe.successThreshold
				if probeType == "liveness" and success ~= nil and tonumber(success) ~= nil and tonumber(success) ~= 1 then
					check_event(
						"失败",
						"Pod " .. ns .. "/" .. podName .. " 容器 " .. containerName .. " 的 liveness 探针 successThreshold 应为 1，当前为 " .. tostring(success),
						{ namespace = ns, name = podName, container = containerName, probe = probeType, field = "successThreshold", value = success }
					)
				end
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
						check_probe_params(ns, name, cName, c.livenessProbe, "liveness")
						check_probe_params(ns, name, cName, c.readinessProbe, "readiness")
					end
				end
			end

			print("Pod 探针参数合规性检查完成")
		`,
	},
}

// registerBuiltinPodProbeComplianceLuaScripts 注册 Pod 探针配置合规性检查相关内置脚本。
func registerBuiltinPodProbeComplianceLuaScripts() {
	BuiltinLuaScripts = append(BuiltinLuaScripts, builtinLuaScriptsPodProbeCompliance...)
}

// init 初始化并注册 Pod 探针配置合规性检查内置脚本。
func init() {
	klog.V(6).Infof("自动注册 Pod 探针配置合规性检查内置脚本")
	registerBuiltinPodProbeComplianceLuaScripts()
}

