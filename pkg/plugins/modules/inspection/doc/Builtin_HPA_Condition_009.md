# HPA Condition 检查

## 介绍

检查 HorizontalPodAutoscaler 的 Condition 状态，ScalingLimited 为 True 或其他 Condition 为 False 时报警。

## 信息

- ScriptCode: Builtin_HPA_Condition_009
- Kind: HorizontalPodAutoscaler
- Group: autoscaling
- Version: v2
- TimeoutSeconds: 45

## 代码

```lua

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
		
```
