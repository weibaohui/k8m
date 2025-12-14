# 事件转发配置与原理说明

## 概述
- 事件转发用于将 Kubernetes 集群中的 **Warning** 类型事件按配置过滤后推送到 Webhook，实现统一告警与聚合。
- 组件分为两部分：
  - Watcher：监听各集群事件并入队
  - Worker：按批次读取未处理事件、执行过滤与转发、记录处理结果

## 配置入口
- 平台参数（全局）：
  - 路径：界面「配置管理 → 事件转发」
  - 接口：`get:/admin/config/all`、`post:/admin/config/update`
  - 字段：
    - `event_forward_enabled`：总开关
    - `event_worker_process_interval`：处理周期（秒）
    - `event_worker_batch_size`：批处理大小
    - `event_worker_max_retries`：最大重试次数
    - `event_watcher_buffer_size`：Watcher 缓存大小
- 规则配置（按集群与Webhook）：
  - 路径：界面「配置 → 事件转发」
  - 接口：`/admin/event/list`、`/admin/event/save`、`/admin/event/delete/:ids`
  - 字段包含：目标集群、Webhook、命名空间/名称/原因过滤、反选、AI总结等

## 原理流程
1. 事件监听（Watcher）
   - 定时检查已连接集群，未启动事件监听则为其启动
   - 将 **Warning** 类型事件入队保存，供 Worker 后续处理
2. 事件处理（Worker）
   - 周期性批量获取未处理事件（按全局批大小）
   - 按每条转发规则进行过滤与推送（Webhook），成功后标记已处理
3. 配置加载
   - `pkg/eventhandler/config/loader.go` 从数据库加载启用的事件规则
   - 全局参数从 `models.Config` 读取：总开关、处理周期、批大小、重试次数、缓存大小

## 启停与多实例协同
- Leader 节点负责启动/停止 Watcher 与 Worker，并负责参数动态更新：
  - 启动时按 `event_forward_enabled` 决定是否启动事件转发
  - 定期同步平台配置（推荐每 1 分钟），若开关或参数变化则启停/更新
- 推荐入口方法（统一调用）：
  - 文件：`pkg/eventhandler/event.go`
  - 方法：
    - `StartEventForwarding()`：按总开关启动 Watcher/Worker
    - `StopEventForwarding()`：停止 Watcher/Worker
    - `SyncEventForwardingFromConfig()`：按最新配置同步（开关或参数变化时生效）
  - 使用方式：
    - Leader 入口 `OnStartedLeading`：调用 `StartEventForwarding()` 并定时执行 `SyncEventForwardingFromConfig()`
    - Leader 退出 `OnStoppedLeading`：调用 `StopEventForwarding()`

## 参数动态更新
- Worker：
  - 支持动态调整处理周期，无需重启；通过 `UpdateConfig()` 即时生效
- Watcher：
  - 缓存大小改变需要重启 Watcher 以应用新的通道容量

## 前端提示
- 事件转发列表页顶部会提示当前总开关状态：
  - 若关闭：提示规则不会生效，并引导前往全局开关页面
  - 若开启：显示当前生效参数（处理周期、批大小、重试次数、缓存大小）

## 常见问题
- 规则配置已填写但不生效：
  - 检查平台总开关 `event_forward_enabled` 是否开启
  - 检查 Leader 节点是否运行事件转发
- 参数更新后未即时生效：
  - Worker 处理周期会在下一次 Tick 自动应用
  - Watcher 缓存大小需要重启 Watcher 方可应用
