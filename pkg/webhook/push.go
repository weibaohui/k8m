package webhook

import (
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/models"
	"k8s.io/klog/v2"
)

func PushMsgToSingleTarget(msg string, raw string, receiver *models.WebhookReceiver) *SendResult {
	sender, err := getSender(receiver.Platform)
	var results *SendResult
	if err != nil {
		klog.V(6).Infof("[webhook] unknown platform: %s, err: %v", receiver.Platform, err)
		results = &SendResult{Status: "failed", RespBody: err.Error()}
		return results
	}
	stdTarget := getStdTarget(receiver)
	results, err = sender.Send(msg, raw, stdTarget)
	if err != nil {
		results = &SendResult{Status: "failed", RespBody: err.Error()}
	}
	return results
}

func PushMsgToAllTargets(msg string, raw string, receivers []*models.WebhookReceiver) []*SendResult {
	var results []*SendResult
	for _, receiver := range receivers {
		result := PushMsgToSingleTarget(msg, raw, receiver)
		klog.V(6).Infof("Push to [%s] %s ,result= [%v] \n", receiver.Platform, receiver.TargetURL, utils.ToJSON(result))
		results = append(results, result)
	}
	return results
}
