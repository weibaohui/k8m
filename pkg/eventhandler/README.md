# Kubernetes事件处理器

这是一个灵活的Kubernetes事件处理模块，支持用户自定义监听规则，并可进行反向选择。

## 功能特性

- **事件监听**: 监听Kubernetes集群中的事件
- **规则匹配**: 支持命名空间、事件原因、事件类型的白名单匹配
- **反向选择**: 可反向选择不匹配规则的事件进行处理
- **异步处理**: Worker层异步处理事件，支持批量处理和重试机制
- **Webhook推送**: 支持将事件推送到外部系统
- **配置热更新**: 支持配置文件的热更新，无需重启服务
- **多数据库支持**: 支持SQLite、PostgreSQL、MySQL

## 架构设计

### 模块结构

```
pkg/eventhandler/
├── model/          # 数据模型
├── store/          # 数据存储层
├── watcher/        # 事件监听层
├── worker/         # 事件处理层
├── webhook/        # Webhook推送
├── config/         # 配置管理
└── example/        # 示例配置
```

### 处理流程

1. **Watcher监听**: 监听Kubernetes事件，应用规则过滤
2. **事件存储**: 将符合条件的事件存储到数据库
3. **Worker处理**: 异步处理未处理的事件
4. **Webhook推送**: 将处理完成的事件推送到外部系统

## 快速开始

### 1. 配置文件

创建配置文件 `eventhandler.yaml`：

```yaml
enabled: true

database:
  type: sqlite
  dsn: k8s_events.db
  max_conns: 10

watcher:
  enabled: true
  resync_interval: 300
  buffer_size: 1000

worker:
  enabled: true
  batch_size: 100
  process_interval: 5000
  max_retries: 3

rule_config:
  enabled: true
  reverse: false
  namespaces:
    - default
    - kube-system
  reasons:
    - FailedMount
    - ImagePullBackOff
  types:
    - Warning

webhook:
  enabled: false
  url: "https://example.com/webhook"
  method: "POST"
  timeout: 30
  retries: 3
```

### 2. 代码集成

```go
package main

import (
    "context"
    "log"
    
    "github.com/weibaohui/k8m/pkg/eventhandler/config"
    "github.com/weibaohui/k8m/pkg/eventhandler/store"
    "github.com/weibaohui/k8m/pkg/eventhandler/watcher"
    "github.com/weibaohui/k8m/pkg/eventhandler/worker"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
)

func main() {
    // 加载配置
    cfg, err := config.LoadConfigFromFile("eventhandler.yaml")
    if err != nil {
        log.Fatal(err)
    }
    
    // 创建Kubernetes客户端
    k8sConfig, err := rest.InClusterConfig()
    if err != nil {
        log.Fatal(err)
    }
    client, err := kubernetes.NewForConfig(k8sConfig)
    if err != nil {
        log.Fatal(err)
    }
    
    // 创建存储层
    eventStore, err := store.NewStore(cfg.Database.Type, cfg.Database.DSN)
    if err != nil {
        log.Fatal(err)
    }
    
    // 创建Watcher
    eventWatcher := watcher.NewEventWatcher(client, eventStore, cfg)
    
    // 创建Worker
    eventWorker := worker.NewEventWorker(eventStore, cfg)
    
    // 启动服务
    ctx := context.Background()
    
    if err := eventWatcher.Start(); err != nil {
        log.Fatal(err)
    }
    
    if err := eventWorker.Start(); err != nil {
        log.Fatal(err)
    }
    
    // 等待中断信号
    <-ctx.Done()
    
    // 清理资源
    eventWatcher.Stop()
    eventWorker.Stop()
}
```

## 配置详解

### 规则配置

规则配置支持以下匹配条件：

- **namespaces**: 命名空间白名单
- **reasons**: 事件原因白名单
- **types**: 事件类型白名单
- **reverse**: 反向选择开关

#### 正向选择（reverse: false）

只处理匹配所有条件的事件：

```yaml
rule_config:
  reverse: false
  namespaces: ["default"]
  reasons: ["FailedMount"]
  types: ["Warning"]
```

#### 反向选择（reverse: true）

处理不匹配任何条件的事件：

```yaml
rule_config:
  reverse: true
  namespaces: ["kube-system"]
  reasons: ["Scheduled"]
  types: ["Normal"]
```

### 数据库配置

支持多种数据库类型：

- **SQLite**: 轻量级，适合单机部署
- **PostgreSQL**: 企业级，支持高并发
- **MySQL**: 广泛应用，性能稳定

### Webhook配置

支持自定义HTTP请求：

```yaml
webhook:
  enabled: true
  url: "https://your-webhook-url.com/events"
  method: "POST"
  headers:
    Authorization: "Bearer your-token"
    X-Custom-Header: "custom-value"
  timeout: 30
  retries: 3
```

## 高级特性

### 事件聚合

Worker层支持事件聚合，避免重复推送相似事件。

### 限流机制

防止同一事件频繁推送，支持基于时间的限流。

### 重试机制

Webhook推送失败时自动重试，支持指数退避。

### 配置热更新

支持运行时更新规则配置，无需重启服务。

## 监控和调试

### 日志级别

使用 `klog.V(6).Infof` 输出调试信息：

```go
klog.V(6).Infof("事件处理完成: %s", event.EvtKey)
```

### 性能指标

可以通过以下指标监控性能：

- 事件处理延迟
- Webhook推送成功率
- 数据库查询性能
- 内存使用量

## 故障排除

### 常见问题

1. **事件未被处理**
   - 检查规则配置是否正确
   - 确认Watcher是否启用
   - 查看日志了解过滤原因

2. **Webhook推送失败**
   - 检查网络连接
   - 验证Webhook URL是否正确
   - 查看重试日志

3. **数据库连接问题**
   - 检查数据库配置
   - 确认数据库服务正常运行
   - 验证连接字符串格式

### 调试建议

1. 启用详细日志记录
2. 使用模拟Webhook进行测试
3. 检查数据库中的事件记录
4. 验证规则匹配逻辑

## 扩展开发

### 添加新的过滤规则

实现自定义的过滤逻辑：

```go
func (w *EventWorker) shouldFilterEvent(event *model.Event) bool {
    // 添加自定义过滤逻辑
    if event.Message contains "specific-pattern" {
        return true
    }
    return false
}
```

### 集成新的Webhook类型

实现 `WebhookClient` 接口：

```go
type CustomWebhookClient struct {
    // 自定义字段
}

func (c *CustomWebhookClient) Push(event *model.Event) error {
    // 实现推送逻辑
    return nil
}

func (c *CustomWebhookClient) PushBatch(events []*model.Event) error {
    // 实现批量推送逻辑
    return nil
}
```

## 许可证

本项目采用开源许可证，详见项目根目录的LICENSE文件。