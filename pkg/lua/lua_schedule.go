package lua

import (
	"context"
	"fmt"
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/models"
	"k8s.io/klog/v2"
)

// TriggerTypeManual 表示手动触发
const TriggerTypeManual = "manual"

// TriggerTypeCron 表示定时触发
const TriggerTypeCron = "cron"

// StartInspection 启动一次巡检任务，并记录执行及每个脚本的结果到数据库
// scheduleID: 可选，定时任务ID（手动触发时为nil）
// cluster: 目标集群
// triggerType: 触发类型（manual/cron）
// createdBy: 发起人
func StartInspection(ctx context.Context, scheduleID *uint, cluster string) (*models.InspectionRecord, error) {
	klog.V(6).Infof("StartInspection, scheduleID: %v, cluster: %s", scheduleID, cluster)
	// 如果sheduleID 不为空，
	// 从数据库中读取sheduleName
	// TODO 记录完成后，统计巡检结果，存入记录表
	// TODO 更新到巡检计划表，最后巡检结果，最后巡检时间
	var scheduleName string
	if scheduleID != nil {
		schedule := &models.InspectionSchedule{}
		schedule.ID = *scheduleID
		schedule, err := schedule.GetOne(nil)
		if err != nil {
			return nil, fmt.Errorf("根据ID获取巡检任务失败: %w", err)
		}
		scheduleName = schedule.Name
	}

	var triggerType = TriggerTypeManual
	if scheduleID != nil {
		triggerType = TriggerTypeCron
	}
	record := &models.InspectionRecord{
		ScheduleID:   scheduleID,
		ScheduleName: scheduleName,
		Cluster:      cluster,
		TriggerType:  triggerType,
		Status:       "running",
		StartTime:    time.Now(),
	}

	if err := record.Save(nil); err != nil {
		return nil, fmt.Errorf("保存巡检执行记录失败: %w", err)
	}

	// 执行所有巡检脚本
	inspection := NewLuaInspection(cluster)
	results := inspection.Start()

	var scriptResults []models.InspectionScriptResult
	var checkEvents []models.InspectionCheckEvent
	for _, res := range results {
		scriptResults = append(scriptResults, models.InspectionScriptResult{
			RecordID:   record.ID,
			ScriptName: res.Name,
			StartTime:  res.StartTime,
			EndTime:    res.EndTime,
			Output:     res.LuaRunOutput,
			ErrorMsg:   fmt.Sprintf("%v", res.LuaRunError),
		})
		for _, e := range res.Events {
			checkEvents = append(checkEvents, models.InspectionCheckEvent{
				RecordID:   record.ID,
				Status:     e.Status,
				Msg:        e.Msg,
				Extra:      utils.ToJSON(e.Extra),
				ScriptName: e.ScriptName,
				Kind:       e.Kind,
				CheckDesc:  e.CheckDesc,
				Namespace:  e.Namespace,
				Name:       e.Name,
			})
		}
	}
	// 保存脚本运行中产生的事件记录
	_ = dao.GenericBatchSave(nil, checkEvents, 100)
	// 保存脚本本身执行结果
	_ = dao.GenericBatchSave(nil, scriptResults, 100)

	//  更新执行记录
	endTime := time.Now()
	record.Status = "success"
	record.EndTime = &endTime
	_ = record.Save(nil)

	return record, nil
}
