package webhook

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/weibaohui/htpl"
	"k8s.io/klog/v2"
)

type DefaultSender struct{}

// DefaultSender 以通用方式发送 Webhook 请求
// 功能：
// 1) 使用 Receiver.BodyTemplate 作为 htpl 模板渲染请求体
// 2) 使用 POST 方法发送到 Receiver.TargetURL
// 3) 不支持签名功能（签名由各平台专用 Sender 处理）
func (d *DefaultSender) Name() string {
	return "default"
}

// Send 发送消息到自定义 Webhook
// 参数：
// - msg: 原始消息字符串，作为模板上下文中的 summary 字段
// - channel: Webhook 接收端配置（包含模板、签名算法、HTTP 方法等）
// 返回：发送结果与错误
func (d *DefaultSender) Send(msg string, raw string, channel *Channel) (*SendResult, error) {
	// 1. 通过 htpl 渲染模板，构造最终请求体；若模板为空或渲染失败，使用原始 msg
	finalBody := msg
	if channel.BodyTemplate != "" {
		bodyTemplate := strings.ReplaceAll(channel.BodyTemplate, "{{msg}}", "${msg}")
		bodyTemplate = strings.ReplaceAll(bodyTemplate, "{{raw}}", "${raw}")
		eng := htpl.NewEngine()
		tpl, err := eng.ParseString(bodyTemplate)
		if err != nil {
			klog.Errorf("Webhook DefaultSender 模板解析失败: platform=%s target=%s template=%q error=%v", channel.Platform, channel.TargetURL, bodyTemplate, err)
		} else {
			ctx := map[string]any{
				"msg": msg,
				"raw": raw,
			}
			rendered, rErr := tpl.Render(ctx)
			if rErr != nil {
				klog.Errorf("Webhook DefaultSender 模板渲染失败: platform=%s target=%s template=%q error=%v", channel.Platform, channel.TargetURL, bodyTemplate, rErr)
			} else if rendered != "" {
				finalBody = rendered
			}
		}
	}

	// 2. 构造 HTTP 请求

	req, err := http.NewRequest("POST", channel.TargetURL, bytes.NewReader([]byte(finalBody)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	// 使用带日志记录的HTTP客户端
	loggedClient := NewLoggedHTTPClient(5*time.Second, "default", channel.TargetURL)
	resp, webhookLog, err := loggedClient.DoWithLogging(req)
	if err != nil {
		return &SendResult{Status: "failed"}, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	// 记录webhook日志到结果中（可选，用于后续查询）
	_ = webhookLog

	status := "success"
	if resp.StatusCode >= 400 {
		status = "failed"
	}
	return &SendResult{
		Status:     status,
		StatusCode: resp.StatusCode,
		RespBody:   string(body),
	}, nil
}
