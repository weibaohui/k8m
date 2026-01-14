# Pod 合规性检查

## 介绍

检查 Pod 的 Pending、调度失败、CrashLoopBackOff、终止异常、ReadinessProbe 失败等状态。

## 信息

- ScriptCode: Builtin_Pod_020
- Kind: Pod
- Group: 
- Version: v1
- TimeoutSeconds: 120

## 代码

```lua

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
		
```
