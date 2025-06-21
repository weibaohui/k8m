package webhook

import (
	"github.com/weibaohui/k8m/pkg/models"
	"k8s.io/klog/v2"
)

func PushMsgToSingleTarget(msg string, receiver *models.WebhookReceiver) *SendResult {
	sender, err := getSender(receiver.Platform)
	var results *SendResult
	if err != nil {
		klog.V(6).Infof("[webhook] unknown platform: %s, err: %v", receiver.Platform, err)
		results = &SendResult{Status: "failed", RespBody: err.Error()}
		return results
	}
	stdTarget := getStdTarget(receiver)
	results, err = sender.Send(msg, stdTarget)
	if err != nil {
		results = &SendResult{Status: "failed", RespBody: err.Error()}
	}
	return results
}

func PushMsgToAllTargets(msg string, receivers []*models.WebhookReceiver) []*SendResult {
	var results []*SendResult
	for _, receiver := range receivers {
		result := PushMsgToSingleTarget(msg, receiver)
		results = append(results, result)
	}
	for _, result := range results {
		klog.V(6).Infof("Push event result: %v \n", result)
	}
	return results
}
