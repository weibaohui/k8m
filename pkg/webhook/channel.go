package webhook

import (
	"github.com/weibaohui/k8m/pkg/models"
)

// Channel represents a user-defined webhook endpoint.
type Channel struct {
	Platform     string
	TargetURL    string
	BodyTemplate string
	SignSecret   string
}

// NewFeishuChannel 快捷创建飞书 Channel
func NewFeishuChannel(targetURL, signSecret string) *Channel {
	return &Channel{
		Platform:     "feishu",
		TargetURL:    targetURL,
		BodyTemplate: `{"msg_type":"text","content":{"text":"%s"}}`,
		SignSecret:   signSecret,
	}
}

// NewDingtalkChannel 快捷创建钉钉 Channel
func NewDingtalkChannel(targetURL, signSecret string) *Channel {
	return &Channel{
		Platform:     "dingtalk",
		TargetURL:    targetURL,
		BodyTemplate: `{"msgtype":"text","text":{"content":"%s"}}`,
		SignSecret:   signSecret,
	}
}

// NewWechatChannel 快捷创建企业微信 Channel（群机器人）
func NewWechatChannel(targetURL string) *Channel {
	return &Channel{
		Platform:     "wechat",
		TargetURL:    targetURL,
		BodyTemplate: `{"msgtype":"markdown","markdown":{"content":"%s"}}`,
		SignSecret:   "",
	}
}

func getStdTarget(receiver *models.WebhookReceiver) *Channel {
	if receiver.Platform == "feishu" {
		rr := NewFeishuChannel(receiver.TargetURL, receiver.SignSecret)
		return rr
	}
	if receiver.Platform == "dingtalk" {
		rr := NewDingtalkChannel(receiver.TargetURL, receiver.SignSecret)
		return rr
	}
	if receiver.Platform == "wechat" {
		rr := NewWechatChannel(receiver.TargetURL)
		return rr
	}
	// 自定义 default 平台：通用映射
	if receiver.Platform == "default" {

		return &Channel{
			Platform:     "default",
			TargetURL:    receiver.TargetURL,
			BodyTemplate: receiver.BodyTemplate,
			SignSecret:   receiver.SignSecret,
		}
	}
	return nil
}
