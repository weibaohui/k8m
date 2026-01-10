# 插件 函数调用 API 设计模式文档

## 概述

`pkg/plugins/api` 包提供了一套插件能力抽象机制，实现了插件之间的解耦和动态能力注册。通过接口抽象和 No-Op 模式，实现了插件间的松耦合依赖，允许插件在运行期动态注册和注销能力实现。

## 设计模式

### 1. 接口抽象模式

通过定义接口来抽象插件能力，调用方只需依赖接口而不需要知道具体实现。

```go
// AIChat 抽象 AI 聊天能力
type AIChat interface {
    Chat(ctx context.Context, prompt string) (string, error)
    ChatNoHistory(ctx context.Context, prompt string) (string, error)
}

// Webhook 抽象 webhook 能力
type Webhook interface {
    PushMsgToAllTargetByIDs(msg string, raw string, receiverIDs []string) []*SendResult
    GetNamesByIds(ids []string) ([]string, error)
}
```

### 2. No-Op 模式（空对象模式）

为每个接口提供默认的空实现，保证在插件未启用时也不会产生空指针异常。

```go
// noopAIChat 为默认的空实现
type noopAIChat struct{}

func (noopAIChat) Chat(ctx context.Context, prompt string) (string, error) {
    return "AI插件未开启", nil
}

func (noopAIChat) ChatNoHistory(ctx context.Context, prompt string) (string, error) {
    return "AI插件未开启", nil
}
```

### 3. 策略模式 + 原子值存储

使用 `atomic.Value` 存储接口实现，支持运行期动态切换，保证线程安全。

```go
var aiChatVal atomic.Value // 保存 AIChat 实现，始终为非 nil

type aiChatHolder struct {
    chat AIChat
}

// RegisterAI 在运行期注册或切换 AI 能力实现
func RegisterAI(chatImpl AIChat, cfgImpl AIConfig) {
    if chatImpl == nil {
        chatImpl = noopAIChat{}
    }
    if cfgImpl == nil {
        cfgImpl = noopAIConfig{}
    }

    aiChatVal.Store(&aiChatHolder{chat: chatImpl})
    aiConfigVal.Store(&aiConfigHolder{cfg: cfgImpl})
}
```

### 4. 服务定位器模式

通过全局函数提供能力访问接口，调用方无需知道具体实现来源。

```go
// AIChatService 返回当前生效的 AIChat 实现，始终非 nil
func AIChatService() AIChat {
    return aiChatVal.Load().(*aiChatHolder).chat
}

// WebhookService 返回当前生效的 Webhook 实现，始终非 nil
func WebhookService() Webhook {
    return webhookVal.Load().(*webhookHolder).svc
}
```

## 使用场景

### 场景 1：AI 插件能力被其他插件调用

#### Inspection 插件调用 AI 能力进行巡检总结


```go
// generateAISummary 使用AI生成智能汇总
func (s *ScheduleBackground) generateAISummary(ctx context.Context, msg *SummaryMsg) (string, error) {
    prompt := fmt.Sprintf(prompt, customTemplate, utils.ToJSONCompact(msg))

    // 使用统一 AI 能力接口，避免跨插件直接依赖实现
    ai := api.AIChatService()
    summary, err := ai.ChatNoHistory(ctx, prompt)
    if err != nil {
        return "", fmt.Errorf("AI汇总请求失败: %v", err)
    }

    return summary, nil
}
```

**优势**：
- Inspection 插件无需知道 AI 插件的具体实现
- 即使 AI 插件未启用，也能安全调用（返回 "AI插件未开启"）
- 可以在运行期动态切换 AI 实现

#### Doc 控制器调用 AI 能力进行文档翻译


```go
func (cc *Controller) Detail(c *response.Context) {
    detail := &DetailReq{}
    err := c.ShouldBindJSON(&detail)
    if err != nil {
        amis.WriteJsonError(c, err)
    }
    if detail.Description != "" {
        q := fmt.Sprintf("请翻译下面的语句，注意直接给出翻译内容，不要解释。待翻译内如如下：\n\n%s", detail.Description)
        ctxInst := amis.GetContextWithUser(c)
        ai := api.AIChatService()
        if result, err := ai.Chat(ctxInst, q); err == nil {
            detail.Translate = result
        }
    }

    amis.WriteJsonData(c, detail)
}
```

### 场景 2：Webhook 插件能力被其他插件调用

#### Inspection 插件调用 Webhook 能力发送通知
 

```go
// PushToHooksByRecordID 根据巡检记录ID发送webhook通知
func (s *ScheduleBackground) PushToHooksByRecordID(recordID uint) ([]*api.SendResult, error) {
    // 查询webhooks
    webhookIDs, err := models.GetWebhookReceiverIDsByRecordID(recordID)
    if err != nil {
        return nil, fmt.Errorf("查询webhooks失败: %v", err)
    }

    // 获取巡检记录内容
    record := &models.InspectionRecord{}
    summary, resultRaw, failedCount, scheduleID, err := record.GetRecordBothContentById(recordID)
    if err != nil {
        return nil, fmt.Errorf("获取巡检记录id=%d的内容失败: %v", recordID, err)
    }

    // 通过统一 Webhook 能力接口发送
    results := api.WebhookService().PushMsgToAllTargetByIDs(summary, resultRaw, webhookIDs)

    return results, nil
}
```
 
 
## 原理详解

### 1. 线程安全保证

使用 `atomic.Value` 存储接口实现，保证并发读写安全。

```go
var aiChatVal atomic.Value

// Store 操作是原子的
aiChatVal.Store(&aiChatHolder{chat: chatImpl})

// Load 操作是原子的
return aiChatVal.Load().(*aiChatHolder).chat
```

### 2. Holder 结构体包装

使用 holder 结构体包装接口，避免 `atomic.Value` 直接存储接口类型（Go 的 atomic.Value 不能直接存储接口类型）。

```go
type aiChatHolder struct {
    chat AIChat
}

// 存储
aiChatVal.Store(&aiChatHolder{chat: chatImpl})

// 读取
holder := aiChatVal.Load().(*aiChatHolder)
return holder.chat
```

### 3. No-Op 实现保证

所有能力在初始化时都会设置 No-Op 实现，确保调用方永远不会遇到空指针。

```go
func initAINoop() {
    aiChatVal.Store(&aiChatHolder{chat: noopAIChat{}})
    aiConfigVal.Store(&aiConfigHolder{cfg: noopAIConfig{}})
}

// AIChatService 始终返回非 nil 的实现
func AIChatService() AIChat {
    return aiChatVal.Load().(*aiChatHolder).chat
}
```

### 4. 动态切换能力

支持在运行期动态切换能力实现，无需重启服务。

```go
// 注册新实现
api.RegisterAI(newAIImpl, newAIConfigImpl)

// 注销（回退到 No-Op）
api.UnregisterAI()

// 再次注册（切换到另一个实现）
api.RegisterAI(anotherAIImpl, anotherAIConfigImpl)
```
 

## 当前支持的能力

### AI 能力

- **AIChat**: AI 聊天能力
  - `Chat(ctx, prompt)`: 带历史记录的对话
  - `ChatNoHistory(ctx, prompt)`: 不带历史记录的对话

- **AIConfig**: AI 配置能力
  - `AnySelect()`: 是否允许任意选择
  - `FloatingWindow()`: 是否启用浮动窗口

### Webhook 能力

- **Webhook**: Webhook 推送能力
  - `PushMsgToAllTargetByIDs(msg, raw, receiverIDs)`: 批量推送消息
  - `GetNamesByIds(ids)`: 根据 ID 查询名称
 
## 总结

插件 API 设计模式通过接口抽象、No-Op 模式、原子值存储和服务定位器模式，实现了：

- **解耦**: 插件间松耦合，易于维护和扩展
- **安全**: No-Op 实现保证系统稳定性
- **灵活**: 支持运行期动态切换能力
- **可测**: 易于注入 Mock 实现进行测试
 
