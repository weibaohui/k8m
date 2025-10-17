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
		"record_date":        record.EndTime,
		"record_id":          recordID,
		"schedule_id":        record.ScheduleID,
		"cluster":            record.Cluster,
		"total_rules":        totalRules,
		"failed_rules":       failCount,
		"failed_list":        events,
		"ai_enabled":         schedule.AIEnabled,
		"ai_prompt_template": schedule.AIPromptTemplate,
	}
	return result, nil
}

// SummaryByAI 生成AI总结
// 参数：msg 包含巡检数据和AI配置的消息
// 参数：format 自定义格式（已废弃，使用msg中的ai_prompt_template）
func (s *ScheduleBackground) SummaryByAI(ctx context.Context, msg map[string]any, format string) (string, error) {
	summary := ""
	var err error
	
	// 检查是否启用AI总结
	aiEnabled, ok := msg["ai_enabled"].(bool)
	if !ok || !aiEnabled {
		return "", fmt.Errorf("该巡检计划未启用AI总结功能")
	}
	
	// 验证必要的数据
	if  len(msg) == 0 {
		return "", fmt.Errorf("巡检数据为空，无法生成AI总结")
	}
	
	// 获取自定义提示词模板
	customTemplate, _ := msg["ai_prompt_template"].(string)
	
	defaultFormat := `
	请按下面的格式给出汇总：
		检测集群：xxx名称
		执行规则数：x个
		问题数：x个
		时间：月日时间
		总结：简短汇总
	`
	
	// 优先使用巡检计划中配置的自定义模板
	if customTemplate != "" {
		klog.V(6).Infof("使用巡检计划中配置的自定义Prompt模板")
		defaultFormat = customTemplate
	} else if format != "" {
		// 兼容旧的format参数
		klog.V(6).Infof("使用传入的自定义Prompt %s", format)
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
		summary, err = service.ChatService().ChatWithCtx(ctx, prompt)

		if err != nil {
			return "", err
		}

	} else {
		summary = "AI功能未开启"
	}

	return summary, err
}

// SaveSummaryBack 保存AI总结结果到数据库
func (s *ScheduleBackground) SaveSummaryBack(id uint, summary string, summaryErr error) error {
	recordModel := &models.InspectionRecord{}
	record, err := recordModel.GetOne(nil, func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", id)
	})
	if err != nil {
		return fmt.Errorf("未找到对应的巡检记录: %v", err)
	}
	if summaryErr != nil {
		record.AISummaryErr = summaryErr.Error()
	}

	record.AISummary = summary
	err = dao.DB().Model(&record).Select("ai_summary_err", "ai_summary").Updates(record).Error
	if err != nil {
		return fmt.Errorf("保存巡检记录的AI总结失败: %v", err)
	}
	return nil
}

// AutoGenerateSummaryIfEnabled 如果启用了AI总结，则自动生成总结
// 该方法在巡检执行完成后被调用
func (s *ScheduleBackground) AutoGenerateSummaryIfEnabled(recordID uint) {
	// 获取巡检数据和AI配置
	msg, err := s.GetSummaryMsg(recordID)
	if err != nil {
		klog.Errorf("获取巡检记录数据失败: %v", err)
		return
	}

	// 检查是否启用AI总结
	aiEnabled, ok := msg["ai_enabled"].(bool)
	if !ok || !aiEnabled {
		klog.V(6).Infof("巡检记录 %d 未启用AI总结功能，跳过自动生成", recordID)
		return
	}

	// 检查AI服务是否可用
	if !service.AIService().IsEnabled() {
		klog.V(6).Infof("AI服务未启用，跳过自动生成总结")
		// 保存错误信息
		s.SaveSummaryBack(recordID, "", fmt.Errorf("AI服务未启用"))
		return
	}

	klog.V(6).Infof("开始为巡检记录 %d 自动生成AI总结", recordID)

	// 生成AI总结
	summary, summaryErr := s.SummaryByAI(context.Background(), msg, "")

	// 保存总结结果
	err = s.SaveSummaryBack(recordID, summary, summaryErr)
	if err != nil {
		klog.Errorf("保存AI总结失败: %v", err)
	} else {
		klog.V(6).Infof("成功为巡检记录 %d 生成并保存AI总结", recordID)
	}
}
