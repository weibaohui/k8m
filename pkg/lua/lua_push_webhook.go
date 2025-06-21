package lua

import (
	"context"
	"fmt"
	"strings"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/webhooksender"
	"gorm.io/gorm"
)

func (s *ScheduleBackground) SummaryAndPushToHooksByRecordID(ctx context.Context, recordID uint, webhooks string) ([]*webhooksender.SendResult, error) {
	// 查询webhooks
	hookModel := &models.WebhookReceiver{}
	hooks, _, err := hookModel.List(dao.BuildDefaultParams(), func(db *gorm.DB) *gorm.DB {
		return db.Where("id in ?", strings.Split(webhooks, ","))
	})
	if err != nil {
		return nil, fmt.Errorf("查询webhooks失败: %v", err)
	}

	var results []*webhooksender.SendResult
	for _, hook := range hooks {
		if hook.Platform == "feishu" {
			AISummary, err := s.SummaryByAI(ctx, recordID, hook.Template)
			if err != nil {
				continue
			}
			receiver := webhooksender.NewFeishuReceiver(hook.TargetURL, hook.SignSecret)
			ret := webhooksender.PushEvent(AISummary, []*webhooksender.WebhookReceiver{
				receiver,
			})
			results = append(results, ret...)
		}
	}

	return results, nil
}

func (s *ScheduleBackground) PushToHooksByRecordID(ctx context.Context, recordID uint, webhooks string) ([]*webhooksender.SendResult, error) {

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
	hooks, _, err := hookModel.List(dao.BuildDefaultParams(), func(db *gorm.DB) *gorm.DB {
		return db.Where("id in ?", strings.Split(webhooks, ","))
	})
	if err != nil {
		return nil, fmt.Errorf("查询webhooks失败: %v", err)
	}

	results := webhooksender.PushMsgToAllReceiver(record.AISummary, hooks)

	return results, nil
}
