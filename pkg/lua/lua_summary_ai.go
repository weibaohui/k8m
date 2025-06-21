package lua

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/service"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

func (s *ScheduleBackground) GetSummaryMsg(recordID uint) (map[string]any, error) {

	// 1. 查询 InspectionRecord
	recordModel := &models.InspectionRecord{}
	record, err := recordModel.GetOne(nil, func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", recordID)
	})
	if err != nil {
		return nil, fmt.Errorf("未找到对应的巡检记录: %v", err)
	}

	if record.ScheduleID == nil {
		return nil, fmt.Errorf("该巡检记录未关联巡检计划")
	}

	// 2. 查询 InspectionSchedule
	scheduleModel := &models.InspectionSchedule{}
	schedule, err := scheduleModel.GetOne(nil, func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", *record.ScheduleID)
	})
	if err != nil {
		return nil, fmt.Errorf("未找到对应的巡检计划: %v", err)
	}

	// 3. 统计规则数
	scriptCodes := utils.SplitAndTrim(schedule.ScriptCodes, ",")
	totalRules := len(scriptCodes)

	// 4. 统计失败数
	eventModel := &models.InspectionCheckEvent{}
	failCount := 0
	events, _, err := eventModel.List(nil, func(db *gorm.DB) *gorm.DB {
		return db.Where("record_id = ? AND event_status = ?", recordID, constants.LuaEventStatusFailed)
	})

	if err == nil {
		failCount = len(events)
	}

	result := gin.H{
		"record_date":  record.EndTime,
		"record_id":    recordID,
		"schedule_id":  record.ScheduleID,
		"cluster":      record.Cluster,
		"total_rules":  totalRules,
		"failed_rules": failCount,
		"failed_list":  events,
	}
	return result, nil
}

// SummaryByAI
// 参数：format 自定义格式
func (s *ScheduleBackground) SummaryByAI(ctx context.Context, msg map[string]any, format string) (string, error) {
	summary := ""
	defaultFormat := `
	请按下面的格式给出汇总：
		检测集群：xxx名称
		执行规则数：x个
		问题数：x个
		时间：月日时间
		总结：简短汇总
	`
	if format != "" {
		klog.V(6).Infof("使用自定义Prompt %s", format)
		defaultFormat = format
	}
	if service.AIService().IsEnabled() {
		prompt := `下面是k8s集群巡检记录，请你进行总结，字数控制在200字以内。
		%s
		要求：
		1、仅做汇总，不要解释
		2、不需要解决方案。
		3、可以合理使用表情符号。
		以下是JSON格式的巡检结果：
		%s
		`
		prompt = fmt.Sprintf(prompt, defaultFormat, utils.ToJSON(msg))
		summary = service.ChatService().ChatWithCtx(ctx, prompt)

	} else {
		summary = "AI功能未开启"
	}

	return summary, nil
}

func (s *ScheduleBackground) SaveSummaryBack(id uint, summary string) error {
	recordModel := &models.InspectionRecord{}
	record, err := recordModel.GetOne(nil, func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id)
	})
	if err != nil {
		return fmt.Errorf("未找到对应的巡检记录: %v", err)
	}

	record.AISummary = summary
	err = dao.DB().Model(&record).Select("ai_summary").Updates(record).Error
	if err != nil {
		return fmt.Errorf("保存巡检记录的AI总结失败: %v", err)
	}
	return nil
}
