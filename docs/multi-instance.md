# 多实例部署与协同（Leader 选举 + Lease 同步）

K8M 支持同时运行多个实例以提高可靠性与扩展性。在多实例模式下，平台通过两项机制完成协同：

- Leader 选举：确保只有一个实例执行定时任务（如集群巡检、Helm 仓库更新）。
- Lease 同步：为每个连接的集群建立租约对象，驱动所有实例对该集群的本地连接/断开保持一致。

本文基于以下实现参考：`pkg/leader/leader.go`、`pkg/lease/manager.go`、`main.go`。

## 原理概述

- 实例标识：每个 K8M 实例在启动时生成唯一的 `InstanceID`（参考 `utils.GenerateInstanceID()`）。
- Leader 选举（`leader.Run`）：
  - 采用 Kubernetes `LeaseLock` 进行选举，锁名为 `k8m-leader-lock`，命名空间默认自动检测（参考 `utils.DetectNamespace()`）。
  - 默认时序为：`LeaseDuration=60s`、`RenewDeadline=50s`、`RetryPeriod=10s`，仅当有可用的集群配置时参与选举；否则直接本地运行为 Leader（不选举）。
  - 成为 Leader 后启动定时任务：集群巡检（Lua）与 Helm 仓库更新；失去 Leader 后停止这些任务（参考 `main.go`）。
- Lease 同步（`lease.Manager`）：
  - 连接某个集群前，实例调用 `EnsureOnConnect` 在宿主集群的指定命名空间创建一个 `Lease`，用于声明“该集群处于连接状态”。
  - `Lease` 名称格式：`<product>-cluster-<sha1(clusterID)前4字节>`，其中 `<product>` 来自 `--product-name`（参考 `flag.Config.ProductName`）。
  - `Lease` 标签固定包含：`app=k8m`、`type=cluster-sync`、`clusterID=<Base64编码>`；`HolderIdentity` 为当前实例的 `InstanceID`。
  - 所有实例通过共享 Informer 监听该命名空间下的 `Lease` 增删改：
    - 有效 `Lease` 新增/更新 → 非责任实例触发本地 `Connect(clusterID)`。
    - `Lease` 删除 → 所有实例触发本地 `Disconnect(clusterID)`。
  - Leader 以 30s 周期清理过期的 `Lease`，从而统一触发断开（参考 `StartLeaderCleanup`）。

## 行为流程

- 连接：
  - A 实例欲连接某集群 → 调用 `EnsureOnConnect` 创建该集群的 `Lease` 并开始续约循环。
  - 若已有其他实例持有有效 `Lease`，A 实例不会重复创建，日志提示“已连接”，并等待 Watcher 驱动本地连接。
  - 其他实例通过 Watcher 看到有效 `Lease` 新增后执行 `Connect(clusterID)`，从而实现所有实例的本地连接保持一致。
- 断开：
  - 责任实例在本地断开时删除对应 `Lease`（`EnsureOnDisconnect`）。
  - 其他实例收到删除事件后执行 `Disconnect(clusterID)`，一致断开。
- 过期清理：
  - 若责任实例异常退出，其 `Lease` 不再续约；Leader 检测到过期后删除该 `Lease`，驱动所有实例断开，保证状态一致性与自愈。

## 参数与配置

以下参数均可通过启动参数或环境变量设置（参考 `pkg/flag/flag.go`）。

- 宿主集群选择：
  - `--host-cluster-id` 或 `HOST_CLUSTER_ID`：指定用于存储 `Lease` 和进行 Leader 选举的宿主集群 ID（从“多集群管理”复制）。
  - 未指定时回退使用 InCluster（`--in-cluster` / `IN_CLUSTER`）。
- Lease 参数：
  - `--lease-namespace` / `LEASE_NAMESPACE`：`Lease` 命名空间，默认自动检测。
  - `--lease-duration-seconds` / `LEASE_DURATION_SECONDS`：`Lease` 有效时长，默认 `60`。
  - `--lease-renew-interval-seconds` / `LEASE_RENEW_INTERVAL_SECONDS`：续约周期，默认 `20`；若未设置或不合法，自动回退为有效时长的 1/3（不低于 20）。
- 监听与重同步：
  - 代码内默认 `ResyncPeriod=30s`（参考 `main.go` 中传入的 `lease.Options`）。
- 启动连接行为：
  - `--connect-cluster` / `CONNECT_CLUSTER`：程序启动后是否自动连接已发现的集群，默认关闭。
- 选举锁：
  - 锁名固定为 `k8m-leader-lock`，命名空间默认自动检测。需要该命名空间在宿主集群中可访问。

## 部署建议

- 在同一或不同节点运行多个 K8M 实例，确保它们能访问相同的宿主集群。
- 建议所有实例共享数据库，以保证平台管理数据的一致性；连接状态由 `Lease` 同步保证，无需依赖单点。
- 将实例置于负载均衡之后对外提供服务。所有实例都会根据 `Lease` 同步进行本地连接/断开；只有 Leader 执行巡检与 Helm 仓库更新等定时任务。
- 日志建议设置 `LOG_V=6` 以获得更详细的中文日志（系统内部使用 `klog.V(6).Infof` 输出）。

## 示例

环境变量示例（可写入 `.env` 或容器环境）：

```
IN_CLUSTER=true
HOST_CLUSTER_ID=your-cluster-id   # 可选；留空则优先 InCluster
LEASE_NAMESPACE=k8m               # 可选；留空自动检测
LEASE_DURATION_SECONDS=60
LEASE_RENEW_INTERVAL_SECONDS=20
CONNECT_CLUSTER=true
PRODUCT_NAME=K8M
LOG_V=6
```

启动参数示例（仅示例，按需调整）：

```
./k8m \
  --in-cluster \
  --connect-cluster \
  --lease-namespace k8m \
  --lease-duration-seconds 60 \
  --lease-renew-interval-seconds 20 \
  --product-name K8M
```

以上配置即可在多实例部署下，通过 Leader 选举与 `Lease` 同步机制，实现集群连接状态一致、任务单点执行与异常自愈。