# Node 合规性检查

## 介绍

检查 Node 的 Condition 状态，非 Ready/EtcdIsVoter 且状态异常时报警。

## 信息

- ScriptCode: Builtin_Node_019
- Kind: Node
- Group: 
- Version: v1
- TimeoutSeconds: 45

## 代码

```lua

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
		
```
