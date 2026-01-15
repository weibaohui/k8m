# HPA ScaleTargetRef 存在性检查

## 介绍

检查 HorizontalPodAutoscaler 的 ScaleTargetRef 指向的对象是否存在。

## 信息

- ScriptCode: Builtin_HPA_ScaleTargetRef_010
- Kind: HorizontalPodAutoscaler
- Group: autoscaling
- Version: v2
- TimeoutSeconds: 60

## 代码

```lua

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
		
```
