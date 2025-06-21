package webhooksender

import (
	"sync"

	"github.com/weibaohui/k8m/pkg/models"
	"k8s.io/klog/v2"
)

// PushEvent sends the event to a list of receivers (并发推送)
func PushEvent(msg string, receivers []*WebhookReceiver) []*SendResult {
	results := make([]*SendResult, len(receivers))
	var wg sync.WaitGroup
	for i, r := range receivers {
		wg.Add(1)
		go func(idx int, receiver *WebhookReceiver) {
			defer wg.Done()
			results[idx] = PushMsgToSingleReceiver(msg, receiver)
		}(i, r)
	}
	wg.Wait()
	return results
}
func PushMsgToSingleReceiver(msg string, receiver *WebhookReceiver) *SendResult {
	sender, err := GetSender(receiver.Platform)
	var results *SendResult
	if err != nil {
		klog.V(6).Infof("[webhook] unknown platform: %s, err: %v", receiver.Platform, err)
		results = &SendResult{Status: "failed", RespBody: err.Error()}
		return results
	}
	results, err = sender.Send(msg, receiver)
	if err != nil {
		results = &SendResult{Status: "failed", RespBody: err.Error()}
	}
	return results
}

func PushMsgToAllReceiver(summary string, hooks []*models.WebhookReceiver) []*SendResult {
	var receivers []*WebhookReceiver
	for _, hook := range hooks {
		switch hook.Platform {
		case "feishu":
			receiver := NewFeishuReceiver(hook.TargetURL, hook.SignSecret)
			receivers = append(receivers, receiver)
			// 可以在此添加更多平台 case
		}
	}

	// 推送到所有的receiver
	results := PushEvent(summary, receivers)
	for _, result := range results {
		klog.V(6).Infof("Push event result: %v \n", result)
	}
	return results
}
