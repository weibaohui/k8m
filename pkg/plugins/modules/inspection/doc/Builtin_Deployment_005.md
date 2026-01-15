# Deployment 配置检查

## 介绍

分析 Deployment 配置问题

## 信息

- ScriptCode: Builtin_Deployment_005
- Kind: Deployment
- Group: apps
- Version: v1
- TimeoutSeconds: 60

## 代码

```lua

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
		
```
