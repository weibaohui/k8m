package models

import (
	"errors"
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"gorm.io/gorm"
)

// EventForwardSetting 中文函数注释：事件转发总开关与运行参数配置表。
type EventForwardSetting struct {
	ID uint `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`

	EventWorkerProcessInterval int `json:"event_worker_process_interval"`
	EventWorkerBatchSize       int `json:"event_worker_batch_size"`
	EventWorkerMaxRetries      int `json:"event_worker_max_retries"`
	EventWatcherBufferSize     int `json:"event_watcher_buffer_size"`

	CreatedAt time.Time `json:"created_at,omitempty" gorm:"<-:create"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

// TableName 中文函数注释：设置表名。
func (EventForwardSetting) TableName() string {
	return "eventhandler_event_forward_settings"
}

// DefaultEventForwardSetting 中文函数注释：返回默认配置（默认关闭）。
func DefaultEventForwardSetting() *EventForwardSetting {
	return &EventForwardSetting{
		EventWorkerProcessInterval: 10,
		EventWorkerBatchSize:       50,
		EventWorkerMaxRetries:      3,
		EventWatcherBufferSize:     1000,
	}
}

// GetOrCreateEventForwardSetting 中文函数注释：获取事件转发配置；若不存在则写入一条默认记录。
func GetOrCreateEventForwardSetting() (*EventForwardSetting, error) {
	db := dao.DB()
	var s EventForwardSetting
	if err := db.Order("id asc").First(&s).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			def := DefaultEventForwardSetting()
			if cErr := db.Create(def).Error; cErr != nil {
				return nil, cErr
			}
			return def, nil
		}
		return nil, err
	}
	return &s, nil
}

// UpdateEventForwardSetting 中文函数注释：更新事件转发配置（确保只维护一条记录）。
func UpdateEventForwardSetting(in *EventForwardSetting) (*EventForwardSetting, error) {
	if in == nil {
		return nil, nil
	}
	cur, err := GetOrCreateEventForwardSetting()
	if err != nil {
		return nil, err
	}

	cur.EventWorkerProcessInterval = in.EventWorkerProcessInterval
	cur.EventWorkerBatchSize = in.EventWorkerBatchSize
	cur.EventWorkerMaxRetries = in.EventWorkerMaxRetries
	cur.EventWatcherBufferSize = in.EventWatcherBufferSize

	if err := dao.DB().Save(cur).Error; err != nil {
		return nil, err
	}
	return cur, nil
}
