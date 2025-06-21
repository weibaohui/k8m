package inspection

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/lua"
	"github.com/weibaohui/k8m/pkg/models"
	"gorm.io/gorm"
)

func Push(c *gin.Context) {
	recordIDStr := c.Param("id")
	if recordIDStr == "" {
		amis.WriteJsonError(c, fmt.Errorf("缺少 record_id 参数"))
		return
	}
	recordID := utils.ToUInt(recordIDStr)
	if recordID == 0 {
		amis.WriteJsonError(c, fmt.Errorf("record_id 参数无效"))
		return
	}

	// 1. 查询 InspectionRecord
	record := &models.InspectionRecord{}
	record, err := record.GetOne(nil, func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", recordID)
	})
	if err != nil {
		amis.WriteJsonError(c, fmt.Errorf("未找到对应的巡检记录: %v", err))
		return
	}
	// 2. 查询 Schedule，查找webhook
	schedule := &models.InspectionSchedule{}
	schedule, err = schedule.GetOne(nil, func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", record.ScheduleID)
	})
	if err != nil {
		amis.WriteJsonError(c, fmt.Errorf("未找到对应的巡检计划: %v", err))
		return
	}

	sb := lua.ScheduleBackground{}
	results, err := sb.SummaryAndPushToHooksByRecordID(context.Background(), recordID, schedule.Webhooks)
	if err != nil {
		amis.WriteJsonError(c, fmt.Errorf("<UNK>: %v", err))
		return
	}
	amis.WriteJsonData(c, results)
}
