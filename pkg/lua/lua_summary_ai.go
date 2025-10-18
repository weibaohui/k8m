package lua

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

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

// SummaryByAI 生成巡检总结
// 参数：msg 包含巡检数据和AI配置的消息
// 参数：format 自定义格式（已废弃，使用msg中的ai_prompt_template）
// 返回：总结内容和错误信息
func (s *ScheduleBackground) SummaryByAI(ctx context.Context, msg map[string]any) (string, error) {
	// msg 参考格式
	// result := gin.H{
	//		"record_date":        record.EndTime,
	//		"record_id":          recordID,
	//		"schedule_id":        record.ScheduleID,
	//		"cluster":            record.Cluster,
	//		"total_rules":        totalRules,
	//		"failed_rules":       failCount,
	//		"failed_list":        events,
	//		"ai_enabled":         schedule.AIEnabled,
	//		"ai_prompt_template": schedule.AIPromptTemplate,
	//	}

	// 验证必要的数据
	if len(msg) == 0 {
		return "", fmt.Errorf("巡检数据为空，无法生成总结")
	}

	// 第一步：生成基础统计汇总
	basicSummary, err := s.generateBasicSummary(msg)
	if err != nil {
		return "", fmt.Errorf("生成基础汇总失败: %v", err)
	}

	// 第二步：检查是否开启AI汇总
	aiEnabled, ok := msg["ai_enabled"].(bool)
	if !ok || !aiEnabled {
		klog.V(6).Infof("AI汇总未启用，返回基础汇总")
		return basicSummary, nil
	}

	// 检查AI服务是否可用
	if !service.AIService().IsEnabled() {
		klog.V(6).Infof("AI服务未启用，返回基础汇总")
		return basicSummary, nil
	}

	// 使用AI进行汇总
	aiSummary, err := s.generateAISummary(ctx, msg)
	if err != nil {
		klog.Errorf("AI汇总失败，返回基础汇总: %v", err)
		return basicSummary, nil
	}

	return aiSummary, nil
}

// generateBasicSummary 生成基础统计汇总
// 参数：msg 包含巡检数据的消息
// 返回：基础汇总内容和错误信息
func (s *ScheduleBackground) generateBasicSummary(msg map[string]any) (string, error) {
	// 提取基础信息
	cluster, _ := msg["cluster"].(string)
	if cluster == "" {
		cluster = "未知集群"
	}

	totalRules, _ := msg["total_rules"].(int)
	failedRules, _ := msg["failed_rules"].(int)

	// 处理巡检时间
	recordDate := ""
	if date, ok := msg["record_date"]; ok {
		if timePtr, ok := date.(*time.Time); ok && timePtr != nil {
			// 转换为本地时间并格式化为易读格式
			localTime := timePtr.Local()
			recordDate = localTime.Format("2006-01-02 15:04:05")
		} else {
			recordDate = fmt.Sprintf("%v", date)
		}
	}
	if recordDate == "" {
		recordDate = "未知时间"
	}

	// 生成基础汇总
	summary := fmt.Sprintf(`📊 巡检汇总报告

🔍 巡检集群：%s
📋 执行规则：%d条
❌ 失败规则：%d条
⏰ 巡检时间：%s

✅ 巡检完成，共发现 %d 个问题需要关注。`,
		cluster,
		totalRules,
		failedRules,
		recordDate,
		failedRules,
	)

	return summary, nil
}

// generateAISummary 使用AI生成智能汇总
// 参数：ctx 上下文，msg 巡检数据，format 自定义格式
// 返回：AI汇总内容和错误信息
func (s *ScheduleBackground) generateAISummary(ctx context.Context, msg map[string]any) (string, error) {
	// 获取自定义提示词模板
	customTemplate, _ := msg["ai_prompt_template"].(string)
	prompt := `以下是k8s集群巡检记录，请你进行总结。
	
		基本要求：
		1、仅做汇总，不要解释
		2、不需要解决方案。
		3、可以合理使用表情符号。
	
	    附加要求：
		%s
		
		以下是JSON格式的巡检结果：
		%s
		`
	prompt = fmt.Sprintf(customTemplate, utils.ToJSON(msg))

	summary, err := service.ChatService().ChatWithCtx(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("AI汇总请求失败: %v", err)
	}

	return summary, nil
}

// SaveSummaryBack 保存AI总结结果到数据库
// 参数：id 巡检记录ID，summary AI总结内容，summaryErr AI总结错误，resultRaw 原始巡检结果JSON字符串
func (s *ScheduleBackground) SaveSummaryBack(id uint, summary string, summaryErr error, resultRaw string) error {
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
	// 更新原始巡检结果
	if resultRaw != "" {
		record.ResultRaw = resultRaw
	}

	err = dao.DB().Model(&record).Select("ai_summary_err", "ai_summary", "result_raw").Updates(record).Error
	if err != nil {
		return fmt.Errorf("保存巡检记录的AI总结失败: %v", err)
	}
	return nil
}

// AutoGenerateSummary 如果启用了AI总结，则自动生成总结
// 该方法在巡检执行完成后被调用，生成的AI总结会保存到数据库中
// 调用时机：巡检完成后立即调用，在发送webhook之前
// 设计原则：AI总结生成与webhook发送分离，确保职责单一
func (s *ScheduleBackground) AutoGenerateSummary(recordID uint) {
	// 获取巡检数据和AI配置
	msg, err := s.GetSummaryMsg(recordID)
	if err != nil {
		klog.Errorf("获取巡检记录数据失败: %v", err)
		return
	}

	// 将原始巡检结果转换为JSON字符串
	resultRawBytes, err := json.Marshal(msg)
	if err != nil {
		klog.Errorf("序列化原始巡检结果失败: %v", err)
		resultRawBytes = []byte("{}")
	}
	resultRaw := string(resultRawBytes)

	klog.V(6).Infof("开始为巡检记录 %d 自动生成AI总结", recordID)
	// 生成AI总结
	summary, summaryErr := s.SummaryByAI(context.Background(), msg)

	// 保存总结结果和原始巡检结果
	err = s.SaveSummaryBack(recordID, summary, summaryErr, resultRaw)
	if err != nil {
		klog.Errorf("保存AI总结失败: %v", err)
	} else {
		klog.V(6).Infof("成功为巡检记录 %d 生成并保存AI总结", recordID)
	}
}
