package models

import (
	"fmt"

	"github.com/weibaohui/k8m/pkg/plugins/modules/webhook"
	hkmodels "github.com/weibaohui/k8m/pkg/plugins/modules/webhook/models"
	"gorm.io/gorm"
)

// GetWebhookReceiversByRecordID 根据巡检记录ID获取关联的webhook接收器列表
// 该函数负责inspection插件与webhook插件之间的协调：
// 1. 从巡检记录中获取关联的计划ID
// 2. 从巡检计划中获取配置的webhook ID列表
// 3. 调用webhook插件的封装接口查询接收器信息
//
// 设计原则：
// - 将跨插件的协调逻辑放在调用方（inspection插件）
// - webhook插件通过export封装接口对外服务，隐藏SQL等实现细节
// - 这样避免了插件间的循环依赖，职责更加清晰
func GetWebhookReceiversByRecordID(recordID uint) ([]*hkmodels.WebhookReceiver, error) {
	// 1. 查询 InspectionRecord
	record := &InspectionRecord{}
	record, err := record.GetOne(nil, func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", recordID)
	})
	if err != nil {
		return nil, fmt.Errorf("查询巡检记录失败: %v", err)
	}

	// 2. 查询 Schedule，获取webhook ID列表
	schedule := &InspectionSchedule{}
	schedule, err = schedule.GetOne(nil, func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", record.ScheduleID)
	})
	if err != nil {
		return nil, fmt.Errorf("查询巡检计划失败: %v", err)
	}

	// 3. 调用webhook插件的封装接口查询接收器
	receivers, err := webhook.GetReceiversByIds(schedule.Webhooks)
	if err != nil {
		return nil, fmt.Errorf("查询webhook接收器失败: %v", err)
	}
	return receivers, nil
}
