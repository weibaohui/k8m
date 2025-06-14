package inspection

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/models"
	"gorm.io/gorm"
)

// Summary 汇总指定scheduleID下的巡检执行信息
// 展示涉及集群数、每个集群涉及的Kind数量、每个Kind检查次数及错误数
func Summary(c *gin.Context) {
	// 1. 获取scheduleID参数
	scheduleID := c.Param("id")
	if scheduleID == "" {
		amis.WriteJsonError(c, fmt.Errorf("缺少scheduleID参数"))
		return
	}

	// 2. 查询所有该scheduleID下的InspectionRecord，收集recordIDs和集群
	recordModel := &models.InspectionRecord{}
	records, _, err := recordModel.List(nil, func(db *gorm.DB) *gorm.DB {
		return db.Where("schedule_id = ?", scheduleID)
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
	events, _, err := eventModel.List(nil, func(db *gorm.DB) *gorm.DB {
		return db.Where("record_id in ?", recordIDs)
	})
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 4. 聚合统计

	totalClusters := len(clusterSet)
	clusterKindMap := map[string]map[string]int{}    // cluster -> kind -> count
	clusterKindErrMap := map[string]map[string]int{} // cluster -> kind -> error count
	for _, e := range events {
		if _, ok := clusterKindMap[e.Cluster]; !ok {
			clusterKindMap[e.Cluster] = map[string]int{}
			clusterKindErrMap[e.Cluster] = map[string]int{}
		}
		clusterKindMap[e.Cluster][e.Kind]++
		if e.EventStatus != "正常" {
			clusterKindErrMap[e.Cluster][e.Kind]++
		}
	}

	// 5. 构建返回结构
	result := gin.H{
		"total_clusters": totalClusters,
		"clusters":       []gin.H{},
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
			"cluster": cluster,
			"kinds":   kindArr,
		})
	}
	amis.WriteJsonData(c, result)
}
