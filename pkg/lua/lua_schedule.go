package lua

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/kom"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

type ScheduleBackground struct {
}

// TriggerTypeManual 表示手动触发
const TriggerTypeManual = "manual"

// TriggerTypeCron 表示定时触发
const TriggerTypeCron = "cron"

// RunByCluster 启动一次巡检任务，并记录执行及每个脚本的结果到数据库
// scheduleID: 可选，定时任务ID（手动触发时为nil）
// cluster: 目标集群
// triggerType: 触发类型（manual/cron）
func (s *ScheduleBackground) RunByCluster(ctx context.Context, scheduleID *uint, cluster string, triggerType string) (*models.InspectionRecord, error) {
	k := kom.Cluster(cluster)
	if k == nil {
		klog.V(6).Infof("巡检 集群【%s】，但是该集群未连接，尝试连接该集群", cluster)
		service.ClusterService().Connect(cluster)
		k = kom.Cluster(cluster)
		if k == nil {
			klog.Errorf("巡检 集群【%s】，但是该集群未连接，尝试连接该集群失败，跳过执行", cluster)
			return nil, fmt.Errorf("巡检 集群【%s】未连接，尝试连接未成功，跳过执行", cluster)
		}
		klog.V(6).Infof("巡检 集群【%s】，已连接", cluster)
	}

	klog.V(6).Infof("开始巡检, scheduleID: %v, cluster: %s", scheduleID, cluster)
	// 如果scheduleID 不为空，
	// 从数据库中读取scheduleName
	var scheduleName string
	if scheduleID == nil {
		return nil, fmt.Errorf("参数错误，scheduleID不能为空")
	}
	schedule := &models.InspectionSchedule{}
	schedule.ID = *scheduleID
	schedule, err := schedule.GetOne(nil)
	if err != nil {
		return nil, fmt.Errorf("根据ID获取巡检计划失败: %w", err)
	}
	scheduleName = schedule.Name

	// 创建一条执行记录
	record := &models.InspectionRecord{
		ScheduleID:   scheduleID,
		ScheduleName: scheduleName,
		Cluster:      cluster,
		TriggerType:  triggerType,
		Status:       "running",
		StartTime:    time.Now(),
	}

	if err := record.Save(nil); err != nil {
		return nil, fmt.Errorf("保存巡检计划执行记录失败: %w", err)
	}

	// 执行所有巡检脚本
	inspection := NewLuaInspection(schedule, cluster)
	results := inspection.Start()

	var scriptResults []*models.InspectionScriptResult
	var checkEvents []*models.InspectionCheckEvent
	var errorCount int
	for _, res := range results {
		result := models.InspectionScriptResult{
			RecordID:   record.ID,
			ScheduleID: scheduleID,
			ScriptName: res.Name,
			StartTime:  res.StartTime,
			EndTime:    res.EndTime,
			StdOutput:  res.LuaRunOutput,
			Cluster:    cluster,
		}
		if res.LuaRunError != nil {
			result.ErrorMsg = res.LuaRunError.Error()
		}
		scriptResults = append(scriptResults, &result)
		for _, e := range res.Events {
			ce := &models.InspectionCheckEvent{
				RecordID:    record.ID,
				ScheduleID:  scheduleID,
				EventStatus: e.Status,
				EventMsg:    e.Msg,
				Extra:       utils.ToJSON(e.Extra),
				ScriptName:  e.ScriptName,
				Kind:        e.Kind,
				CheckDesc:   e.CheckDesc,
				Namespace:   e.Namespace,
				Name:        e.Name,
				Cluster:     cluster,
			}
			if s.IsEventStatusPass(e.Status) {
				ce.EventStatus = string(constants.LuaEventStatusNormal) // 统一状态描述为正常
			} else {
				errorCount += 1
				ce.EventStatus = string(constants.LuaEventStatusFailed)
			}
			checkEvents = append(checkEvents, ce)

		}
	}
	// 保存脚本运行中产生的事件记录
	_ = dao.GenericBatchSave(nil, checkEvents, 100)
	// 保存脚本本身执行结果
	_ = dao.GenericBatchSave(nil, scriptResults, 100)

	// 统计错误数

	//  更新执行记录
	endTime := time.Now()
	record.Status = "success"
	record.EndTime = &endTime
	record.ErrorCount = errorCount

	_ = record.Save(nil)

	klog.V(6).Infof("集群巡检完成。集群巡检记录ID=%d", record.ID)

	// 更新集群巡检计划运行结果
	schedule.LastRunTime = &endTime
	schedule.ErrorCount = errorCount
	_ = schedule.Save(nil, func(db *gorm.DB) *gorm.DB {
		return db.Select("last_run_time", "error_count")
	})

	// 自动生成总结，包括使用AI
	s.AutoGenerateSummary(record.ID)

	// 发送webhook通知
	go func() {
		_, _ = s.PushToHooksByRecordID(record.ID)
	}()

	return record, nil
}

var localCron *cron.Cron
var once sync.Once

func InitClusterInspection() {
	once.Do(func() {
		localCron = cron.New()
		localCron.Start()
		klog.V(6).Infof("集群巡检启动")
		sb := ScheduleBackground{}
		sb.StartFromDB()
	})
}

// IsEventStatusPass 判断事件状态是否为通过
// 这里的通过状态包括：正常、pass、ok、success、通过
// 入库前将状态描述文字统一为正常、失败两种
func (s *ScheduleBackground) IsEventStatusPass(status string) bool {
	return status == "正常" || status == "pass" || status == "ok" || status == "success" || status == "通过"
}

// StartFromDB  后台自动执行调度
func (s *ScheduleBackground) StartFromDB() {

	// 1、读取数据库中的定义，然后创建
	sch := models.InspectionSchedule{}

	list, _, err := sch.List(nil, func(db *gorm.DB) *gorm.DB {
		return db.Where("enabled is true")
	})
	if err != nil {
		klog.Errorf("读取定时任务失败%v", err)
		return
	}
	var count int
	for _, schedule := range list {
		s.Add(schedule.ID)
		count += 1
	}
	klog.V(6).Infof("启动集群巡检任务完成，共启动%d个", count)
}
func (s *ScheduleBackground) Remove(scheduleID uint) {
	sch := models.InspectionSchedule{}
	sch.ID = scheduleID
	item, err := sch.GetOne(nil, func(db *gorm.DB) *gorm.DB {
		return db
	})
	if err != nil {
		klog.Errorf("读取定时任务[id=%d]失败  %v", scheduleID, err)
		return
	}
	if item.CronRunID != 0 {
		// 首先标记任务为已删除状态，防止正在执行的任务继续运行
		taskControlManager.MarkTaskDeleted(scheduleID)

		// 从cron调度器中移除任务
		localCron.Remove(item.CronRunID)

		// 清理任务控制信息
		taskControlManager.RemoveTask(scheduleID)

		// 清空数据库中的CronRunID
		item.CronRunID = 0
		_ = item.Save(nil)
	}
	klog.V(6).Infof("移除集群定时巡检任务[id=%d]", scheduleID)

}

// Add 添加一个新的定时巡检任务到cron调度器中
// 该方法会从数据库读取指定的巡检计划，并创建对应的定时任务
//
// 重要：为了避免Go闭包变量捕获问题，该方法会创建局部变量副本
// 这确保每个定时任务都有自己独立的数据副本，不会被后续的Add调用影响
// 如果不这样做，多个巡检计划可能会错误地使用相同的集群配置
//
// 参数:
//
//	scheduleID - 要添加的巡检计划ID
func (s *ScheduleBackground) Add(scheduleID uint) {
	// 1、读取数据库中的定义，然后创建
	sch := models.InspectionSchedule{}
	sch.ID = scheduleID
	item, err := sch.GetOne(nil, func(db *gorm.DB) *gorm.DB {
		return db.Where("enabled is true")
	})
	if err != nil {
		klog.Errorf("读取定时任务[id=%d]失败  %v", scheduleID, err)
		return
	}
	// 先清除，再添加执行
	if item.CronRunID != 0 && localCron != nil {
		localCron.Remove(item.CronRunID)
		// 同时清理任务控制信息
		taskControlManager.RemoveTask(scheduleID)
	}
	if item.Cron != "" && localCron != nil {
		klog.V(6).Infof("注册定时任务item: %s", item.Cron)
		// 注册定时任务
		// 创建局部副本以避免闭包变量捕获问题
		// 这样确保每个定时任务都有自己独立的数据副本，不会被后续的Add调用影响
		scheduleIDCopy := item.ID
		clustersCopy := item.Clusters
		cronExpr := item.Cron

		// 在AddFunc执行前生成随机字符串
		randomToken := utils.RandNLengthString(8)

		entryID, err := localCron.AddFunc(cronExpr, func() {
			// 在任务执行前检查是否已被删除
			if taskControlManager.IsTaskDeleted(randomToken) {
				klog.V(6).Infof("任务已被删除，跳过执行: scheduleID=%d, token=%s", scheduleIDCopy, randomToken)
				return
			}

			klog.V(6).Infof("开始执行定时任务: scheduleID=%d, token=%s", scheduleIDCopy, randomToken)
			for _, cluster := range strings.Split(clustersCopy, ",") {
				// 在每个集群执行前再次检查任务是否已被删除
				if taskControlManager.IsTaskDeleted(randomToken) {
					klog.V(6).Infof("任务在执行过程中被删除，停止执行: scheduleID=%d, cluster=%s, token=%s", scheduleIDCopy, cluster, randomToken)
					break
				}
				_, _ = s.RunByCluster(context.Background(), &scheduleIDCopy, cluster, TriggerTypeCron)
			}
		})
		if err != nil {
			klog.Errorf("定时任务注册失败%v", err)
			return
		}

		// 注册任务控制信息，建立随机字符串与EntryID的映射关系
		taskControlManager.RegisterTask(scheduleID, entryID, randomToken)

		// 更新EntryID
		item.CronRunID = entryID
		_ = item.Save(nil)
	}
	klog.V(6).Infof("启动集群定时巡检任务[id=%d]", scheduleID)

}
