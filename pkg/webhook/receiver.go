package webhook

import (
	"fmt"

	"github.com/weibaohui/k8m/pkg/models"
)

// Receiver represents a user-defined webhook endpoint.
type Receiver struct {
	Platform      string
	TargetURL     string
	Method        string
	Headers       map[string]string
	Template      string
	SignSecret    string
	SignAlgo      string // e.g. "hmac-sha256", "feishu"
	SignHeaderKey string // e.g. "X-Signature" or unused
}

// NewFeishuReceiver 快捷创建飞书 Receiver
func NewFeishuReceiver(targetURL, signSecret string) *Receiver {
	return &Receiver{
		Platform:      "feishu",
		TargetURL:     targetURL,
		Method:        "POST",
		Headers:       map[string]string{},
		Template:      `{"msg_type":"text","content":{"text":"%s"}}`,
		SignSecret:    signSecret,
		SignAlgo:      "feishu",
		SignHeaderKey: "", // 飞书不需要 header 签名，是 URL 参数
	}
}

// Validate 校验 Receiver 配置合法性
func (r *Receiver) Validate() error {
	if r.Platform == "" {
		return fmt.Errorf("platform is required")
	}
	if r.TargetURL == "" {
		return fmt.Errorf("target url is required")
	}
	if r.Method == "" {
		return fmt.Errorf("http method is required")
	}
	if r.Template == "" {
		return fmt.Errorf("template is required")
	}
	return nil
}

func getStdTarget(receiver *models.WebhookReceiver) *Receiver {
	if receiver.Platform == "feishu" {
		rr := NewFeishuReceiver(receiver.TargetURL, receiver.SignSecret)
		return rr
	}
	return nil
}
