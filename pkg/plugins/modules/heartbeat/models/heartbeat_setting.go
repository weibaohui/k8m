package models

import (
	"errors"

	"github.com/weibaohui/k8m/internal/dao"
	"gorm.io/gorm"
)

type HeartbeatSetting struct {
	ID                          uint `gorm:"primaryKey;autoIncrement" json:"id,omitempty"`
	HeartbeatIntervalSeconds    int  `gorm:"default:30" json:"heartbeat_interval_seconds"`       // 心跳间隔时间（秒）
	HeartbeatFailureThreshold   int  `gorm:"default:3" json:"heartbeat_failure_threshold"`       // 心跳失败阈值
	ReconnectMaxIntervalSeconds int  `gorm:"default:3600" json:"reconnect_max_interval_seconds"` // 重连最大间隔时间（秒）
	MaxRetryAttempts            int  `gorm:"default:100" json:"max_retry_attempts"`              // 最大重试次数，默认100次
}

func (HeartbeatSetting) TableName() string {
	return "heartbeat_settings"
}

func DefaultHeartbeatSetting() *HeartbeatSetting {
	return &HeartbeatSetting{
		HeartbeatIntervalSeconds:    30,
		HeartbeatFailureThreshold:   3,
		ReconnectMaxIntervalSeconds: 3600,
		MaxRetryAttempts:            100,
	}
}

func GetOrCreateHeartbeatSetting() (*HeartbeatSetting, error) {
	db := dao.DB()
	var s HeartbeatSetting
	if err := db.Order("id asc").First(&s).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			def := DefaultHeartbeatSetting()
			if cErr := db.Create(def).Error; cErr != nil {
				return nil, cErr
			}
			return def, nil
		}
		return nil, err
	}
	return &s, nil
}

func UpdateHeartbeatSetting(in *HeartbeatSetting) (*HeartbeatSetting, error) {
	if in == nil {
		return nil, nil
	}
	cur, err := GetOrCreateHeartbeatSetting()
	if err != nil {
		return nil, err
	}

	cur.HeartbeatIntervalSeconds = in.HeartbeatIntervalSeconds
	cur.HeartbeatFailureThreshold = in.HeartbeatFailureThreshold
	cur.ReconnectMaxIntervalSeconds = in.ReconnectMaxIntervalSeconds
	cur.MaxRetryAttempts = in.MaxRetryAttempts

	if err := dao.DB().Save(cur).Error; err != nil {
		return nil, err
	}
	return cur, nil
}
