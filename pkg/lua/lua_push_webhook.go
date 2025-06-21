package lua

import (
	"context"
	"fmt"
	"strings"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/webhook"
	"gorm.io/gorm"
)

func (s *ScheduleBackground) SummaryAndPushToHooksByRecordID(ctx context.Context, recordID uint, webhookIDs string) ([]*webhook.SendResult, error) {
	// 查询webhooks
	hookModel := &models.WebhookReceiver{}
	receivers, _, err := hookModel.List(dao.BuildDefaultParams(), func(db *gorm.DB) *gorm.DB {
		return db.Where("id in ?", strings.Split(webhookIDs, ","))
	})
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

func (s *ScheduleBackground) PushToHooksByRecordID(ctx context.Context, recordID uint, webhookIDs string) ([]*webhook.SendResult, error) {

	// 1. 查询 InspectionRecord
	record := &models.InspectionRecord{}
	record, err := record.GetOne(nil, func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", recordID)
	})
	if err != nil {
		return nil, fmt.Errorf("未找到对应的巡检记录: %v", err)
	}

	// 查询webhooks
	hookModel := &models.WebhookReceiver{}
	receivers, _, err := hookModel.List(dao.BuildDefaultParams(), func(db *gorm.DB) *gorm.DB {
		return db.Where("id in ?", strings.Split(webhookIDs, ","))
	})
	if err != nil {
		return nil, fmt.Errorf("查询webhooks失败: %v", err)
	}

	results := webhook.PushMsgToAllTargets(record.AISummary, receivers)

	return results, nil
}
