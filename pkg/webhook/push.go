package webhook

import (
	"context"

	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/models"
	"k8s.io/klog/v2"
)

// Global webhook client instance
var defaultClient = NewWebhookClient()

// PushMsgToSingleTarget sends a message to a single webhook receiver using the new architecture.
func PushMsgToSingleTarget(msg string, raw string, receiver *models.WebhookReceiver) *SendResult {
	config := NewWebhookConfig(receiver)

	// Use the new WebhookClient
	result, err := defaultClient.Send(context.Background(), msg, raw, config)
	if err != nil {
		klog.Errorf("[webhook] Failed to send to [%s] %s: %v",
			receiver.Platform, receiver.TargetURL, err)
		if result == nil {
			result = &SendResult{
				Status:   "failed",
				RespBody: err.Error(),
				Error:    err,
			}
		}
	}

	klog.V(6).Infof("[webhook] Push to [%s] %s, result=[%v]",
		receiver.Platform, receiver.TargetURL, utils.ToJSON(result))

	return result
}

// PushMsgToAllTargets sends a message to multiple webhook receivers.
func PushMsgToAllTargets(msg string, raw string, receivers []*models.WebhookReceiver) []*SendResult {
	var results []*SendResult
	for _, receiver := range receivers {
		result := PushMsgToSingleTarget(msg, raw, receiver)
		results = append(results, result)
	}
	return results
}
