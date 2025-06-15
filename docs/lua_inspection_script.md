# Lua 脚本与 `lua_kubectl` 用法说明

本项目支持通过 Lua 脚本对 Kubernetes 资源进行灵活检测和操作。`lua_kubectl` 提供了丰富的链式方法，便于在 Lua 脚本中以声明式方式查询、筛选、获取和文档化 K8s 资源。

## 一、如何编写 Lua 检测脚本

1. **入口对象**：
   - 脚本中通过 `kubectl` 对象进行所有操作。
   - 支持链式调用，便于组合多种查询条件。

2. **基本流程**：
   - 通过 `GVK` 指定资源类型。
   - 可选地设置命名空间、名称、标签选择器等。
   - 调用 `List` 或 `Get` 获取资源。
   - 可用 `Doc` 获取字段文档说明。

### 示例脚本

```lua
-- 查询所有 default 命名空间下的 Deployment 资源
local deployments, err = kubectl:GVK("apps", "v1", "Deployment"):Namespace("default"):List()
if err then
    print("查询出错：", err)
else
    for i, d in ipairs(deployments) do
        print(d.metadata.name)
    end
end

-- 获取指定名称的 Pod
local pod, err = kubectl:GVK("", "v1", "Pod"):Namespace("kube-system"):Name("coredns-xxxx"):Get()
if err then
    print("获取出错：", err)
else
    print(pod.metadata)
end


-- 获取指定 Deployment 的副本数文档
local doc, err = kubectl:GVK("apps", "v1", "Deployment"):Cache(10):Doc("spec.replicas")
if err then
    print( "获取 Deployment Doc 失败".. tostring(err))
	return
end
print("Deployment Doc 获取成功: " .. doc)

-- 查询所有命名空间下带有特定标签的 Service
local svcs, err = kubectl:GVK("", "v1", "Service"):AllNamespace():WithLabelSelector("app=nginx"):List()

-- 检查所有命名空间下 data 和 binaryData 都为空的 ConfigMap
local configmaps, err = kubectl:GVK("", "v1", "ConfigMap"):AllNamespace():List()
if err then
    print("获取 ConfigMap 失败" .. tostring(err))
    return
end
for _, cm in ipairs(configmaps) do
    local cmName = cm.metadata.name
    local cmNamespace = cm.metadata.namespace
    local isEmpty = true
    if cm.data then
        for k, v in pairs(cm.data) do
            isEmpty = false
            break
        end
    end
    if isEmpty and cm.binaryData then
        for k, v in pairs(cm.binaryData) do
            isEmpty = false
            break
        end
    end
    if isEmpty then
        check_event("失败", "[空数据] ConfigMap " .. cmNamespace .. "/" .. cmName .. " 的 data 和 binaryData 字段都为空", {namespace=cmNamespace, name=cmName})
    end
end
print("ConfigMap 空数据检测完成")
```

## 二、可用方法与说明

### 1. `GVK(group, version, kind)`
- 说明：指定资源的 Group、Version、Kind。
- 返回：新的 kubectl 实例。
- 示例：`kubectl:GVK("apps", "v1", "Deployment")`

### 2. `Namespace(ns)`
- 说明：设置命名空间。
- 示例：`:Namespace("default")`

### 3. `Name(name)`
- 说明：设置资源名称。
- 示例：`:Name("my-deploy")`

### 4. `WithLabelSelector(selector)`
- 说明：设置标签选择器，筛选资源。
- 示例：`:WithLabelSelector("app=nginx")`

### 5. `AllNamespace()`
- 说明：查询所有命名空间下的资源。
- 示例：`:AllNamespace()`

### 6. `Cache(seconds)`
- 说明：设置缓存时间，单位为秒。适合频繁查询场景。
- 示例：`:Cache(30)`

### 7. `List()`
- 说明：获取资源列表。
- 返回：Lua 表（数组），每个元素为资源对象。
- 示例：`:List()`

### 8. `Get()`
- 说明：获取单个资源。
- 返回：Lua 表（对象）。
- 示例：`:Get()`

### 9. `Doc(fieldPath)`
- 说明：获取指定字段的文档说明。
- 参数：如 `"spec.replicas"`。
- 返回：Lua 表（文档内容）。
- 示例：`:Doc("spec.replicas")`

### 10. `check_event(status, msg, extra?)`
- 说明：用于在 Lua 检测脚本中上报结构化检测事件，便于巡检系统收集、展示和统计异常或告警信息。
- 参数：
  - `status` (string)：事件状态，通常为 `失败`（失败）。
  - `msg` (string)：事件描述信息。
  - `extra` (table，可选)：附加信息表，支持自定义字段，常用如 `name`（资源名）、`namespace`（命名空间）等。
- 返回：无返回值。
- 示例：
```lua
-- 检查某资源副本数是否异常
if deploy.spec.replicas ~= deploy.status.replicas then
    check_event("失败", "副本数不一致", {name=deploy.metadata.name, namespace=deploy.metadata.namespace})
end
```
- 典型用法：
  - 在检测逻辑中发现失败等情况时调用。
  - 支持多次调用，所有事件会被系统收集并展示在巡检报告中。

## 三、错误处理

所有方法调用返回值均为 `(结果, 错误信息)`，如无错误则错误信息为 `nil`。

## 四、进阶用法

- 方法可链式组合，顺序不限。
- 支持自定义缓存、标签、命名空间等多条件组合。
- 适合用于自定义资源检测、合规性校验、批量查询等场景。

## 五、AI Prompt：让大模型帮你生成检测规则

如果你不会编写 Lua 检测脚本，可以通过向大模型（如 ChatGPT、Copilot、通义千问等）提问，自动生成所需的规则脚本。你可以参考以下 Prompt 模板：

**通用 Prompt 模板：**


请帮我用 Lua 语言，基于如下链式 API，编写一个 Kubernetes 资源检测脚本：
-- 入口对象为 kubectl，支持 GVK、Namespace、AllNamespace、WithLabelSelector、List、Get、Doc、check_event 等方法。
-- 目标：检测所有命名空间下副本数不一致的 Deployment，正常情况下应该是deployment.spec.replicas == deployment.status.replicas。

要求：
1、返回完整 Lua 代码。
2、检测到异常时，必须调用 check_event("失败", "描述信息", {name=资源名, namespace=命名空间})这样的方法。

如果你需要检测特定资源的副本数一致性
以下是你可以参考示例脚本（Deployment 副本数一致性检测脚本）：
```lua
local deployments, err = kubectl:GVK("apps", "v1", "Deployment"):Cache(360):AllNamespace():List()
if err then
    print("获取 Deployment 失败" .. tostring(err))
    return
end
for _, deploy in ipairs(deployments) do
    local specReplicas = deploy.spec and deploy.spec.replicas or 1
    local statusReplicas = deploy.status and deploy.status.replicas or 0
    if specReplicas ~= statusReplicas then
        check_event("失败", "副本数不一致", {name=deploy.metadata.name, namespace=deploy.metadata.namespace})
    end
end
print("Deployment 副本数一致性检测完成")
```
