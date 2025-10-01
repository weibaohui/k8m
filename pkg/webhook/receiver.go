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
	BodyTemplate  string
	SignSecret    string
	SignAlgo      string // e.g. "hmac-sha256", "feishu", "dingtalk"
	SignHeaderKey string // e.g. "X-Signature" or unused
}

// NewFeishuReceiver 快捷创建飞书 Receiver
func NewFeishuReceiver(targetURL, signSecret string) *Receiver {
	return &Receiver{
		Platform:      "feishu",
		TargetURL:     targetURL,
		Method:        "POST",
		Headers:       map[string]string{},
		BodyTemplate:  `{"msg_type":"text","content":{"text":"%s"}}`,
		SignSecret:    signSecret,
		SignAlgo:      "feishu",
		SignHeaderKey: "", // 飞书不需要 header 签名，是 URL 参数
	}
}

// NewDingtalkReceiver 快捷创建钉钉 Receiver
func NewDingtalkReceiver(targetURL, signSecret string) *Receiver {
	return &Receiver{
		Platform:      "dingtalk",
		TargetURL:     targetURL,
		Method:        "POST",
		Headers:       map[string]string{},
		BodyTemplate:  `{"msgtype":"text","text":{"content":"%s"}}`,
		SignSecret:    signSecret,
		SignAlgo:      "dingtalk",
		SignHeaderKey: "", // 钉钉不需要 header 签名，是 URL 参数
	}
}

// NewWechatReceiver 快捷创建企业微信 Receiver（群机器人）
func NewWechatReceiver(targetURL string) *Receiver {
	return &Receiver{
		Platform:      "wechat",
		TargetURL:     targetURL,
		Method:        "POST",
		Headers:       map[string]string{},
		BodyTemplate:  `{"msgtype":"markdown","markdown":{"content":"%s"}}`,
		SignSecret:    "",
		SignAlgo:      "",
		SignHeaderKey: "",
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
	if r.BodyTemplate == "" {
		return fmt.Errorf("template is required")
	}
	return nil
}

func getStdTarget(receiver *models.WebhookReceiver) *Receiver {
	if receiver.Platform == "feishu" {
		rr := NewFeishuReceiver(receiver.TargetURL, receiver.SignSecret)
		return rr
	}
	if receiver.Platform == "dingtalk" {
		rr := NewDingtalkReceiver(receiver.TargetURL, receiver.SignSecret)
		return rr
	}
	if receiver.Platform == "wechat" {
		rr := NewWechatReceiver(receiver.TargetURL)
		return rr
	}
	// 自定义 default 平台：通用映射
	if receiver.Platform == "default" {
		method := receiver.Method
		if method == "" {
			method = "POST"
		}
		return &Receiver{
			Platform:      "default",
			TargetURL:     receiver.TargetURL,
			Method:        method,
			Headers:       map[string]string{},
			BodyTemplate:  receiver.Template,
			SignSecret:    receiver.SignSecret,
			SignAlgo:      receiver.SignAlgo,
			SignHeaderKey: receiver.SignHeaderKey,
		}
	}
	return nil
}
