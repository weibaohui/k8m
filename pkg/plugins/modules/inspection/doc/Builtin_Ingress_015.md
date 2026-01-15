# Ingress 合规性检查

## 介绍

检查 Ingress 是否指定 IngressClass，引用的 IngressClass/Service/Secret 是否存在。

## 信息

- ScriptCode: Builtin_Ingress_015
- Kind: Ingress
- Group: networking
- Version: v1
- TimeoutSeconds: 75

## 代码

```lua

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
		
```
