package models

import (
	"fmt"
	"strings"

	"github.com/weibaohui/k8m/internal/dao"
	hkmodels "github.com/weibaohui/k8m/pkg/plugins/modules/webhook/models"
	"gorm.io/gorm"
)

// GetWebhookReceiversByRecordID 根据巡检记录ID获取关联的webhook接收器列表
// 该函数负责inspection插件与webhook插件之间的协调：
// 1. 从巡检记录中获取关联的计划ID
// 2. 从巡检计划中获取配置的webhook ID列表
// 3. 调用webhook插件查询具体的接收器信息
//
// 设计原则：
// - 将跨插件的协调逻辑放在调用方（inspection插件）
// - webhook插件只负责基础的CRUD操作，不依赖inspection插件
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

	// 检查 webhooks 字段是否为空
	if strings.TrimSpace(schedule.Webhooks) == "" {
		return []*hkmodels.WebhookReceiver{}, nil
	}

	// 3. 调用webhook插件查询接收器（只通过ID列表查询，不涉及inspection的业务逻辑）
	receiver := &hkmodels.WebhookReceiver{}
	receivers, _, err := receiver.List(dao.BuildDefaultParams(), func(db *gorm.DB) *gorm.DB {
		return db.Where("id in ?", strings.Split(schedule.Webhooks, ","))
	})
	if err != nil {
		return nil, fmt.Errorf("查询webhook接收器失败: %v", err)
	}
	return receivers, nil
}
