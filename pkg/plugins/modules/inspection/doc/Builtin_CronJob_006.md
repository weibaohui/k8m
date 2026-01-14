# CronJob 合规性检查

## 介绍

检查 CronJob 是否被挂起、调度表达式是否合法、startingDeadlineSeconds 是否为负数

## 信息

- ScriptCode: Builtin_CronJob_006
- Kind: CronJob
- Group: 
- Version: v1
- TimeoutSeconds: 45

## 代码

```lua

			-- 内置 Cron 表达式基本校验（Kubernetes 使用标准 5 字段）
			local function split_fields(expr)
				local fields = {}
				for token in string.gmatch(expr or "", "%S+") do table.insert(fields, token) end
				return fields
			end
			local function validate_part(part, min, max, allow_names)
				if part == "*" then return true end
				local step = string.match(part, "^%*/(%d+)$")
				if step then return tonumber(step) and tonumber(step) >= 1 end
				local a,b = string.match(part, "^(%d+)%-(%d+)$")
				if a and b then a=tonumber(a); b=tonumber(b); return a and b and a>=min and b<=max and a<=b end
				local num = tonumber(part)
				if num and num>=min and num<=max then return true end
				if allow_names and string.match(part, "^[A-Za-z]+$") then return true end
				return false
			end
			local function validate_field(field, min, max, allow_names)
				for part in string.gmatch(field, "[^,]+") do
					if not validate_part(part, min, max, allow_names) then return false end
				end
				return true
			end
			local function is_valid_cron(expr)
				if not expr or expr == "" then return false, "表达式为空" end
				if string.match(expr, "^@%w+$") then return true end -- 支持 @yearly 等描述符
				local f = split_fields(expr)
				if #f ~= 5 then return false, "字段数不是5" end
				if not validate_field(f[1], 0, 59, false) then return false, "分钟字段非法" end
				if not validate_field(f[2], 0, 23, false) then return false, "小时字段非法" end
				if not validate_field(f[3], 1, 31, false) then return false, "日字段非法" end
				if not validate_field(f[4], 1, 12, true) then return false, "月字段非法" end
				if not validate_field(f[5], 0, 7, true) then return false, "周字段非法" end
				return true
			end
			local cronjobs, err = kubectl:GVK("batch", "v1", "CronJob"):AllNamespace(""):List()
			if err then
				print("获取 CronJob 失败: " .. tostring(err))
				return
			end
			local doc_suspend, _ = kubectl:GVK("batch", "v1", "CronJob"):Doc("spec.suspend")
			local doc_schedule, _ = kubectl:GVK("batch", "v1", "CronJob"):Doc("spec.schedule")
			local doc_deadline, _ = kubectl:GVK("batch", "v1", "CronJob"):Doc("spec.startingDeadlineSeconds")
			for _, cj in ipairs(cronjobs) do
				local ns = cj.metadata and cj.metadata.namespace or "default"
				local name = cj.metadata and cj.metadata.name or ""
				-- 检查挂起
				if cj.spec and cj.spec.suspend == true then
					check_event("失败", "CronJob " .. name .. " 已被挂起", {namespace=ns, name=name, doc=doc_suspend})
				end
				-- 检查调度表达式合法性（5 字段）
				if cj.spec and cj.spec.schedule ~= nil then
					local ok, reason = is_valid_cron(cj.spec.schedule)
					if not ok then
						check_event("失败", "CronJob " .. name .. " 的调度表达式非法: " .. tostring(reason), {namespace=ns, name=name, value=cj.spec.schedule, doc=doc_schedule})
					end
				end
				-- 检查 startingDeadlineSeconds
				if cj.spec and cj.spec.startingDeadlineSeconds ~= nil then
					if tonumber(cj.spec.startingDeadlineSeconds) < 0 then
						check_event("失败", "CronJob " .. name .. " 的 startingDeadlineSeconds 为负数", {namespace=ns, name=name, value=cj.spec.startingDeadlineSeconds, doc=doc_deadline})
					end
				end
			end
			print("CronJob 合规性检查完成")
		
```
