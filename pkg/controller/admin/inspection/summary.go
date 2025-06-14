package inspection

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	"gorm.io/gorm"
)

// Summary 汇总指定scheduleID下的巡检执行信息
// 展示涉及集群数、每个集群涉及的Kind数量、每个Kind检查次数及错误数
// Summary 统计巡检计划执行情况，支持按时间范围过滤
// @param start_time 可选，起始时间（格式：2006-01-02T15:04:05Z07:00）
// @param end_time 可选，结束时间（格式：2006-01-02T15:04:05Z07:00）
func Summary(c *gin.Context) {
	params := dao.BuildParams(c)
	params.PerPage = 100000
	// 1. 获取scheduleID参数
	scheduleID := c.Param("id")

	// 新增：解析时间范围参数
	var startTime, endTime time.Time
	var err error
	startTimeStr := c.Param("start_time")
	endTimeStr := c.Param("end_time")
	if startTimeStr != "" {
		startTime, err = time.Parse(time.RFC3339, startTimeStr)
		if err != nil {
			amis.WriteJsonError(c, fmt.Errorf("start_time 格式错误，应为 RFC3339"))
			return
		}
	}
	if endTimeStr != "" {
		endTime, err = time.Parse(time.RFC3339, endTimeStr)
		if err != nil {
			amis.WriteJsonError(c, fmt.Errorf("end_time 格式错误，应为 RFC3339"))
			return
		}
	}

	// 2. 查询所有该scheduleID下的InspectionRecord，收集recordIDs和集群
	recordModel := &models.InspectionRecord{}
	records, _, err := recordModel.List(params, func(db *gorm.DB) *gorm.DB {
		query := db
		if scheduleID != "" {
			query = query.Where("schedule_id = ?", scheduleID)
		}
		if !startTime.IsZero() {
			query = query.Where("created_at >= ?", startTime)
		}
		if !endTime.IsZero() {
			query = query.Where("created_at <= ?", endTime)
		}
		return query.Order("id desc")
	})
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	if len(records) == 0 {
		amis.WriteJsonData(c, gin.H{"summary": "无执行记录"})
		return
	}
	recordIDs := make([]uint, 0, len(records))
	clusterSet := map[string]struct{}{}
	for _, r := range records {
		recordIDs = append(recordIDs, r.ID)
		clusterSet[r.Cluster] = struct{}{}
	}

	// 3. 查询所有相关InspectionCheckEvent
	eventModel := &models.InspectionCheckEvent{}
	events, _, err := eventModel.List(params, func(db *gorm.DB) *gorm.DB {
		return db.Where("record_id in ?", recordIDs)
	})
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 4. 聚合统计

	totalClusters := len(clusterSet)
	totalRuns := len(records) // 巡检计划执行次数

	clusterKindMap := map[string]map[string]int{}    // cluster -> kind -> count
	clusterKindErrMap := map[string]map[string]int{} // cluster -> kind -> error count
	for _, e := range events {
		if _, ok := clusterKindMap[e.Cluster]; !ok {
			clusterKindMap[e.Cluster] = map[string]int{}
			clusterKindErrMap[e.Cluster] = map[string]int{}
		}
		clusterKindMap[e.Cluster][e.Kind]++
		if !isEventStatusPass(e.EventStatus) {
			clusterKindErrMap[e.Cluster][e.Kind]++

		}

	}

	// 5. 构建返回结构
	result := gin.H{
		"total_clusters": totalClusters,
		"total_runs":     totalRuns, // 新增字段：执行次数
		"clusters":       []gin.H{},
	}
	// 新增：如果 scheduleID 为空，增加运行巡检计划数
	if scheduleID == "" {
		var count int64
		dao.DB().Model(&models.InspectionRecord{}).Distinct("schedule_id").Count(&count)
		result["total_schedules"] = count
	}
	// 统计每个集群的执行次数
	clusterRunCount := map[string]int{}
	for _, r := range records {
		clusterRunCount[r.Cluster]++
	}
	for cluster, kindMap := range clusterKindMap {
		var kindArr []gin.H
		for kind, count := range kindMap {
			errCount := clusterKindErrMap[cluster][kind]
			kindArr = append(kindArr, gin.H{
				"kind":        kind,
				"count":       count,
				"error_count": errCount,
			})
		}
		result["clusters"] = append(result["clusters"].([]gin.H), gin.H{
			"cluster":   cluster,
			"run_count": clusterRunCount[cluster], // 新增字段：该集群执行次数
			"kinds":     kindArr,
		})
	}
	// 新增：统计最新一次执行情况
	var latestRun gin.H
	if len(records) > 0 {
		latestRecord := records[0]
		kindStatus := map[string]map[string]int{} // kind -> status -> count
		for _, e := range events {
			if e.RecordID == latestRecord.ID {
				if _, ok := kindStatus[e.Kind]; !ok {
					kindStatus[e.Kind] = map[string]int{"pass": 0, "fail": 0}
				}
				if isEventStatusPass(e.EventStatus) {
					kindStatus[e.Kind]["pass"]++
				} else {
					kindStatus[e.Kind]["fail"]++
				}
			}
		}
		var kindArr []gin.H
		for kind, statusMap := range kindStatus {
			kindArr = append(kindArr, gin.H{
				"kind":         kind,
				"normal_count": statusMap["pass"],
				"error_count":  statusMap["fail"],
			})
		}
		latestRun = gin.H{
			"record_id": latestRecord.ID,
			"run_time":  latestRecord.CreatedAt,
			"kinds":     kindArr,
		}
		result["latest_run"] = latestRun
	}
	amis.WriteJsonData(c, result)
}

// isEventStatusPass 判断巡检事件状态是否为通过
func isEventStatusPass(status string) bool {
	return status == "正常" || status == "pass" || status == "ok" || status == "success" || status == "通过"
}
