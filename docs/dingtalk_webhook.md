# 钉钉群机器人Webhook集成

本文档介绍了如何在系统中使用钉钉群机器人Webhook功能。

## 钉钉机器人安全设置

钉钉群机器人支持多种安全设置：
1. 自定义关键词
2. IP地址限制
3. 加签（推荐）

其中加签方式提供了最高的安全性，通过双向认证确保消息来源的合法性。

## 加签算法

钉钉机器人的加签算法如下：

1. 把 timestamp+"\n"+ 密钥当做签名字符串
2. 使用HmacSHA256算法计算签名
3. 然后进行Base64 encode
4. 最后把签名参数再进行urlEncode，得到最终的签名

## 使用方法

### 1. 创建钉钉机器人

在钉钉群中创建自定义机器人，并选择"加签"安全设置，获取加签密钥。

### 2. 配置Webhook接收器

```go
// 创建钉钉接收器
receiver := webhook.NewDingtalkReceiver(
    "https://oapi.dingtalk.com/robot/send?access_token=YOUR_ACCESS_TOKEN",
    "SECyour_secret_here", // 加签密钥
)
```

### 3. 发送消息

```go
// 获取钉钉发送器
sender, err := webhook.GetSender("dingtalk")
if err != nil {
    log.Fatalf("获取钉钉发送器失败: %v", err)
}

// 发送消息
result, err := sender.Send("这是一条测试消息", receiver)
if err != nil {
    log.Fatalf("发送消息失败: %v", err)
}
```

## 配置参数说明

- `TargetURL`: 钉钉机器人的Webhook地址，包含access_token参数
- `SignSecret`: 加签密钥，以"SEC"开头的字符串
- `Platform`: 平台类型，固定为"dingtalk"
- `SignAlgo`: 签名算法，固定为"dingtalk"

## 注意事项

1. 时间戳使用毫秒级精度
2. 签名需要进行Base64编码后再进行URL编码
3. 消息格式遵循钉钉机器人要求的JSON格式