package lua

import (
	"context"
	"fmt"

	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/webhook"
)

// SummaryAndPushToHooksByRecordID 为指定巡检记录生成AI总结并推送到所有关联的webhook
// 现在使用 schedule 中的 AIPromptTemplate 而不是 webhook 中的 Template 字段
// 
// 已废弃：该方法将AI总结生成和webhook发送耦合在一起，不推荐使用
// 推荐使用：先调用 AutoGenerateSummaryIfEnabled() 生成AI总结，再调用 PushToHooksByRecordID() 发送webhook
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
		// 使用 schedule 中的 AIPromptTemplate，不再使用 receiver.Template
		// SummaryByAI 方法会自动从 msg 中获取 ai_prompt_template
		AISummary, summaryErr := s.SummaryByAI(ctx, msg, "")

		_ = s.SaveSummaryBack(recordID, AISummary, summaryErr)

		ret := webhook.PushMsgToSingleTarget(AISummary, receiver)
		results = append(results, ret)

	}
	return results, nil
}

// PushToHooksByRecordID 根据巡检记录ID发送webhook通知
// 该方法从数据库中获取已生成的AI总结，然后发送到所有关联的webhook
// 调用时机：在AutoGenerateSummaryIfEnabled()完成后调用
// 设计原则：单纯的webhook发送功能，不负责AI总结生成
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
