# HPA 资源配置检查

## 介绍

检查 HPA 关联对象的 Pod 模板中所有容器是否配置了 requests 和 limits。

## 信息

- ScriptCode: Builtin_HPA_Resource_011
- Kind: HorizontalPodAutoscaler
- Group: autoscaling
- Version: v2
- TimeoutSeconds: 75

## 代码

```lua

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
		
```
