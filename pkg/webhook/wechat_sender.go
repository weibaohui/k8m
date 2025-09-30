package webhook

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

// WechatSender 企业微信群机器人 Sender，实现 webhook 发送能力。
// 参考官方企业微信群机器人接口，默认以 markdown 格式发送消息。
// 注意：企业微信群机器人仅需在 URL 中携带 key，无需额外签名。
type WechatSender struct{}

// Name 返回平台名称，用于注册与选择对应 Sender。
func (w *WechatSender) Name() string {
	return "wechat"
}

// Send 发送消息到企业微信群机器人。
// 参数 msg 为最终要发送的文本内容；receiver 为统一的目标配置。
// 默认使用 markdown 消息体结构：
// {
//   "msgtype": "markdown",
//   "markdown": {
//     "content": "..."
//   }
// }
func (w *WechatSender) Send(msg string, receiver *Receiver) (*SendResult, error) {
	finalURL := receiver.TargetURL

	// 构造企业微信 markdown 消息体
	payload := map[string]any{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"content": msg,
		},
	}

	msgBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", finalURL, bytes.NewReader(msgBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range receiver.Headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return &SendResult{Status: "failed"}, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	status := "success"
	if resp.StatusCode >= 400 {
		status = "failed"
	}

	return &SendResult{
		Status:     status,
		StatusCode: resp.StatusCode,
		RespBody:   string(respBody),
	}, nil
}