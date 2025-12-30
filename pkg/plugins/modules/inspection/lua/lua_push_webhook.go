package lua

import (
	"fmt"

	"github.com/weibaohui/k8m/pkg/plugins/modules/inspection/models"
	"github.com/weibaohui/k8m/pkg/plugins/modules/webhook"
	hkmodels "github.com/weibaohui/k8m/pkg/plugins/modules/webhook/models"

	"k8s.io/klog/v2"
)

// PushToHooksByRecordID 根据巡检记录ID发送webhook通知
// 该方法从数据库中获取已生成的AI总结，然后发送到所有关联的webhook
// 调用时机：在AutoGenerateSummaryIfEnabled()完成后调用
// 设计原则：单纯的webhook发送功能，不负责AI总结生成
func (s *ScheduleBackground) PushToHooksByRecordID(recordID uint) ([]*webhook.SendResult, error) {

	// 查询webhooks
	receiver := &hkmodels.WebhookReceiver{}
	receivers, err := receiver.ListByRecordID(recordID)
	if err != nil {
		return nil, fmt.Errorf("查询webhooks失败: %v", err)
	}
	record := &models.InspectionRecord{}
	summary, resultRaw, failedCount, scheduleID, err := record.GetRecordBothContentById(recordID)
	if err != nil {
		return nil, fmt.Errorf("获取巡检记录id=%d的内容失败: %v", recordID, err)
	}

	// 通过failedCount==0时，检查计划中的开关配置，是否开启跳过0失败的条目。
	if failedCount == 0 {
		klog.V(6).Infof("巡检记录id=%d失败项数为0", recordID)
		schedule := &models.InspectionSchedule{}
		// 如果跳过0失败的条目
		if schedule.CheckSkipZeroFailedCount(scheduleID) {
			klog.V(4).Infof("巡检计划id=%d配置了跳过0失败的条目[巡检记录id=%d]，不发送webhook", *scheduleID, recordID)
			return nil, nil
		}
	}

	results := webhook.PushMsgToAllTargets(summary, resultRaw, receivers)

	return results, nil
}
