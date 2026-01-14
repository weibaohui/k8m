# Job 合规性检查

## 介绍

检查 Job 是否被挂起（suspend）以及是否有失败（status.failed > 0）

## 信息

- ScriptCode: Builtin_Job_016
- Kind: Job
- Group: batch
- Version: v1
- TimeoutSeconds: 45

## 代码

```lua

			local jobs, err = kubectl:GVK("batch", "v1", "Job"):AllNamespace(""):List()
			if err then print("获取 Job 失败: " .. tostring(err)) return end
			for _, job in ipairs(jobs) do
				if job.spec and job.spec.suspend == true then
					check_event("失败", "Job " .. job.metadata.name .. " 已被挂起", {namespace=job.metadata.namespace, name=job.metadata.name})
				end
				if job.status and job.status.failed and job.status.failed > 0 then
					check_event("失败", "Job " .. job.metadata.name .. " 有失败记录 (failed=" .. tostring(job.status.failed) .. ")", {namespace=job.metadata.namespace, name=job.metadata.name, failed=job.status.failed})
				end
			end
			print("Job 合规性检查完成")
		
```
