package webhook

import (
	"github.com/weibaohui/k8m/pkg/models"
)

// Receiver represents a user-defined webhook endpoint.
type Receiver struct {
	Platform     string
	TargetURL    string
	BodyTemplate string
	SignSecret   string
}

// NewFeishuReceiver 快捷创建飞书 Receiver
func NewFeishuReceiver(targetURL, signSecret string) *Receiver {
	return &Receiver{
		Platform:     "feishu",
		TargetURL:    targetURL,
		BodyTemplate: `{"msg_type":"text","content":{"text":"%s"}}`,
		SignSecret:   signSecret,
	}
}

// NewDingtalkReceiver 快捷创建钉钉 Receiver
func NewDingtalkReceiver(targetURL, signSecret string) *Receiver {
	return &Receiver{
		Platform:     "dingtalk",
		TargetURL:    targetURL,
		BodyTemplate: `{"msgtype":"text","text":{"content":"%s"}}`,
		SignSecret:   signSecret,
	}
}

// NewWechatReceiver 快捷创建企业微信 Receiver（群机器人）
func NewWechatReceiver(targetURL string) *Receiver {
	return &Receiver{
		Platform:     "wechat",
		TargetURL:    targetURL,
		BodyTemplate: `{"msgtype":"markdown","markdown":{"content":"%s"}}`,
		SignSecret:   "",
	}
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

		return &Receiver{
			Platform:     "default",
			TargetURL:    receiver.TargetURL,
			BodyTemplate: receiver.BodyTemplate,
			SignSecret:   receiver.SignSecret,
		}
	}
	return nil
}
