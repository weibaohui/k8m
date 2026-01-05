package core

import (
	"context"

	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/plugins/modules/webhook/models"
	"k8s.io/klog/v2"
)

// Global webhook client instance
var defaultClient = NewWebhookClient()

// PushMsgToSingleTarget sends a message to a single webhook receiver using the new architecture.
func PushMsgToSingleTarget(msg string, raw string, receiver *models.WebhookReceiver) *SendResult {
	if receiver == nil {
		klog.Errorf("[webhook] nil receiver")
		return &SendResult{Status: "failed", Error: ErrInvalidConfig}
	}
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

	klog.V(8).Infof("[webhook] Push to [%s] %s, result=[%v]",
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

// PushMsgToAllTargetByIDs sends a message to multiple webhook receivers.
func PushMsgToAllTargetByIDs(msg string, raw string, receiverIDs []string) []*SendResult {
	var results []*SendResult
	//根据ID 获取所有receiver
	m := models.WebhookReceiver{}
	receivers, err := m.GetReceiversByIds(receiverIDs)
	if err != nil {
		klog.Errorf("[webhook] Failed to get receivers by ids %v: %v", receiverIDs, err)
		return results
	}
	for _, receiver := range receivers {
		result := PushMsgToSingleTarget(msg, raw, receiver)
		results = append(results, result)
	}
	return results
}
