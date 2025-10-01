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
// - receiver: Webhook 接收端配置（包含模板、签名算法、HTTP 方法等）
// 返回：发送结果与错误
func (d *DefaultSender) Send(msg string, receiver *Receiver) (*SendResult, error) {
	// 1. 通过 htpl 渲染模板，构造最终请求体；若模板为空或渲染失败，使用原始 msg
	finalBody := msg
	if receiver.BodyTemplate != "" {
		bodyTemplate := strings.ReplaceAll(receiver.BodyTemplate, "{{msg}}", "${msg}")
		eng := htpl.NewEngine()
		tpl, err := eng.ParseString(bodyTemplate)
		if err != nil {
			klog.Errorf("Webhook DefaultSender 模板解析失败: platform=%s target=%s template=%q error=%v", receiver.Platform, receiver.TargetURL, bodyTemplate, err)
		} else {
			ctx := map[string]any{
				"msg": msg,
			}
			rendered, rErr := tpl.Render(ctx)
			if rErr != nil {
				klog.Errorf("Webhook DefaultSender 模板渲染失败: platform=%s target=%s template=%q error=%v", receiver.Platform, receiver.TargetURL, bodyTemplate, rErr)
			} else if rendered != "" {
				finalBody = rendered
			}
		}
	}

	// 2. 构造 HTTP 请求

	req, err := http.NewRequest("POST", receiver.TargetURL, bytes.NewReader([]byte(finalBody)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	// 为方便调试，将请求体打印到日志
	klog.V(8).Infof(" Sending POST request to %s with body: %s", receiver.TargetURL, finalBody)

	// 4. 发送请求
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return &SendResult{Status: "failed"}, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

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
