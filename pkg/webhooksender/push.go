package webhooksender

import (
	"sync"

	"k8s.io/klog/v2"
)

// PushEvent sends the event to a list of receivers (并发推送)
func PushEvent(event *InspectionCheckEvent, receivers []*WebhookReceiver) []SendResult {
	results := make([]SendResult, len(receivers))
	var wg sync.WaitGroup
	for i, r := range receivers {
		wg.Add(1)
		go func(idx int, receiver *WebhookReceiver) {
			defer wg.Done()
			sender, err := GetSender(receiver.Platform)
			if err != nil {
				klog.V(6).Infof("[webhook] unknown platform: %s, err: %v", receiver.Platform, err)
				results[idx] = SendResult{Status: "failed", RespBody: err.Error()}
				return
			}
			res, err := sender.Send(event, receiver)
			if err != nil {
				results[idx] = SendResult{Status: "failed", RespBody: err.Error()}
			} else {
				results[idx] = *res
			}
		}(i, r)
	}
	wg.Wait()
	return results
}
