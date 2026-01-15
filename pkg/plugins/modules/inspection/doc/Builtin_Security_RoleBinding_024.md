# Security RoleBinding 通配符检测

## 介绍

检测 RoleBinding 关联的 Role 是否包含通配符权限。

## 信息

- ScriptCode: Builtin_Security_RoleBinding_024
- Kind: RoleBinding
- Group: rbac.authorization.k8s.io
- Version: v1
- TimeoutSeconds: 75

## 代码

```lua

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
		
```
