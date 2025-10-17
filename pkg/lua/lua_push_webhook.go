package lua

import (
	"fmt"

	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/webhook"
)

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
	summary, _, err := record.GetRecordContentById(recordID)
	if err != nil {
		return nil, fmt.Errorf("获取巡检记录id=%d的内容失败: %v", recordID, err)
	}

	results := webhook.PushMsgToAllTargets(summary, receivers)

	return results, nil
}
