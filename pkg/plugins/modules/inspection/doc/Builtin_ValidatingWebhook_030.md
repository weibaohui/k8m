# ValidatingWebhookConfiguration 合规性检查

## 介绍

检查 ValidatingWebhookConfiguration 的 webhook 指向的 Service 是否存在、是否有活跃 Pod、Pod 状态。

## 信息

- ScriptCode: Builtin_ValidatingWebhook_030
- Kind: ValidatingWebhookConfiguration
- Group: admissionregistration.k8s.io
- Version: v1
- TimeoutSeconds: 0

## 代码

```lua

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
		
```
