package lua

import (
	"context"
	"fmt"

	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/webhook"
)

func (s *ScheduleBackground) SummaryAndPushToHooksByRecordID(ctx context.Context, recordID uint) ([]*webhook.SendResult, error) {
	// 查询webhooks
	receiver := &models.WebhookReceiver{}
	receivers, err := receiver.ListByRecordID(recordID)
	if err != nil {
		return nil, fmt.Errorf("查询webhooks失败: %v", err)
	}

	msg, err := s.GetSummaryMsg(recordID)
	if err != nil {
		return nil, err
	}

	var results []*webhook.SendResult
	for _, receiver := range receivers {
		AISummary, err := s.SummaryByAI(ctx, msg, receiver.Template)
		if err != nil {
			AISummary += err.Error()
		}
		_ = s.SaveSummaryBack(recordID, AISummary)

		ret := webhook.PushMsgToSingleTarget(AISummary, receiver)
		results = append(results, ret)

	}
	return results, nil
}

func (s *ScheduleBackground) PushToHooksByRecordID(recordID uint) ([]*webhook.SendResult, error) {

	// 查询webhooks
	receiver := &models.WebhookReceiver{}
	receivers, err := receiver.ListByRecordID(recordID)
	if err != nil {
		return nil, fmt.Errorf("查询webhooks失败: %v", err)
	}
	record := &models.InspectionRecord{}
	summary, err := record.GetAISummaryById(recordID)
	if err != nil {
		return nil, fmt.Errorf("获取巡检记录id=%d的AI总结失败", recordID)
	}

	results := webhook.PushMsgToAllTargets(summary, receivers)

	return results, nil
}
