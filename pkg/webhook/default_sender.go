package webhook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"time"

	"github.com/weibaohui/htpl"
	"k8s.io/klog/v2"
)

type DefaultSender struct{}

// DefaultSender 以通用方式发送 Webhook 请求，支持可选的 HMAC-SHA256 签名
// 功能：
// 1) 使用 Receiver.BodyTemplate 作为 htpl 模板渲染请求体
// 2) 使用 Receiver.Method/TargetURL 作为 HTTP 方法与地址
// 3) 当 SignAlgo=="hmac-sha256" 时，按 SignSecret 计算签名并写入 SignHeaderKey
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
		eng := htpl.NewEngine()
		tpl, err := eng.ParseString(receiver.BodyTemplate)
		if err == nil {
			ctx := map[string]any{
				"summary": msg,
				"time":    time.Now().Format(time.RFC3339),
			}
			if rendered, rErr := tpl.Render(ctx); rErr == nil && rendered != "" {
				finalBody = rendered
			}
		}
	}

	// 2. 构造 HTTP 请求
	method := receiver.Method
	if method == "" {
		method = "POST"
	}
	req, err := http.NewRequest(method, receiver.TargetURL, bytes.NewReader([]byte(finalBody)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range receiver.Headers {
		req.Header.Set(k, v)
	}

	// 3. 可选签名：HMAC-SHA256
	if receiver.SignAlgo == "hmac-sha256" && receiver.SignSecret != "" && receiver.SignHeaderKey != "" {
		h := hmac.New(sha256.New, []byte(receiver.SignSecret))
		h.Write([]byte(finalBody))
		signature := hex.EncodeToString(h.Sum(nil))
		req.Header.Set(receiver.SignHeaderKey, signature)
	}

	// 为方便调试，将请求体打印到日志
	klog.V(6).Infof("Sending %s request to %s with body: %s", method, receiver.TargetURL, finalBody)

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
