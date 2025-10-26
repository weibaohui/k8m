# Webhook 模块重构总结

## 概述

本次重构解决了原有 webhook 模块设计中的多个问题，提供了更清晰、更易扩展的架构。

## 主要改进

### 1. 架构重新设计

**原有架构问题：**
- `Channel` 和 `Sender` 职责混乱
- 代码重复，硬编码严重
- 扩展性差，添加新平台需要修改多处代码

**新架构：**
```
WebhookConfig (配置层) → PlatformAdapter (平台适配层) → WebhookClient (传输层)
```

### 2. 核心组件

#### WebhookConfig
- 替代原有的 `Channel`
- 统一管理 webhook 配置信息
- 提供配置验证和默认模板功能

#### PlatformAdapter 接口
```go
type PlatformAdapter interface {
    Name() string
    FormatMessage(msg, raw string, config *WebhookConfig) ([]byte, error)
    SignRequest(url string, body []byte, secret string) (string, error)
    GetContentType() string
}
```

#### WebhookClient
- 统一的 HTTP 传输层
- 处理完整的发送流程：验证 → 格式化 → 签名 → 发送 → 解析响应

### 3. 平台适配器

实现了以下平台适配器：
- **DingtalkAdapter**: 钉钉 webhook 支持
- **FeishuAdapter**: 飞书 webhook 支持  
- **WechatAdapter**: 企业微信 webhook 支持
- **DefaultAdapter**: 通用 webhook 支持，支持自定义模板

### 4. 向后兼容

- 保留了 `RegisterAllSenders` 函数用于 API 兼容性
- 新架构完全替代了旧的实现

## 使用示例

### 新架构使用方式

```go
// 创建配置
config := &WebhookConfig{
    Platform:  "dingtalk",
    TargetURL: "https://oapi.dingtalk.com/robot/send?access_token=xxx",
    SignSecret: "your-secret",
}

// 创建客户端并发送
client := NewWebhookClient()
result, err := client.Send(context.Background(), "Hello World", "raw data", config)
```

### 简化的发送方式

```go
// 直接使用 WebhookReceiver
receiver := &models.WebhookReceiver{
    Platform:  "feishu",
    TargetURL: "https://open.feishu.cn/open-apis/bot/v2/hook/xxx",
}

result := PushMsgToSingleTarget("Hello World", "raw data", receiver)
```

## 文件结构

```
pkg/webhook/
├── adapters.go      # 平台适配器实现
├── client.go        # WebhookClient 传输层
├── config.go        # WebhookConfig 配置层
├── errors.go        # 错误定义
├── init.go          # 初始化和注册
├── push.go          # 高级发送接口
├── types.go         # 类型定义和接口
└── webhook_test.go  # 测试文件
```

## 测试

运行测试验证功能：

```bash
go test -v ./pkg/webhook/
```

## 迁移指南

### 对于新代码
推荐使用新的 `WebhookConfig` + `WebhookClient` 架构。

### 对于现有代码
现有代码无需修改，原有的 `PushMsgToSingleTarget` 和 `PushMsgToAllTargets` 函数已经内部使用新架构，但保持了相同的接口。

### 添加新平台
1. 实现 `PlatformAdapter` 接口
2. 在 `init.go` 中注册新适配器
3. 在 `WebhookConfig.Validate()` 中添加平台验证

## 优势总结

1. **职责清晰**: 配置、适配、传输三层分离
2. **易于扩展**: 添加新平台只需实现 `PlatformAdapter`
3. **代码复用**: 统一的传输层和错误处理
4. **配置驱动**: 支持模板和验证
5. **向后兼容**: 现有代码无需修改
6. **易于测试**: 每个组件都可以独立测试