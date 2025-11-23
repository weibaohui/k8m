package models

import (
	"fmt"
	"time"
)

// K8sEvent 事件处理器使用的K8s事件模型
type K8sEvent struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	EvtKey    string    `gorm:"uniqueIndex;not null" json:"evt_key"`
	Type      string    `gorm:"type:varchar(16);not null" json:"type"`
	Reason    string    `gorm:"type:varchar(128);not null" json:"reason"`
	Level     string    `gorm:"type:varchar(16);not null" json:"level"`
	Namespace string    `gorm:"type:varchar(64);not null;index" json:"namespace"`
	Name      string    `gorm:"type:varchar(128);not null" json:"name"`
	Message   string    `gorm:"type:text;not null" json:"message"`
	Timestamp time.Time `gorm:"not null;index" json:"timestamp"`
	Processed bool      `gorm:"default:false;index" json:"processed"`
	Attempts  int       `gorm:"default:0" json:"attempts"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// TableName 设置表名
func (K8sEvent) TableName() string {
	return "k8s_events"
}

// IsWarning 判断是否为警告事件
func (e *K8sEvent) IsWarning() bool {
	return e.Type == "Warning" || e.Level == "warning"
}

// GenerateEvtKey 生成事件键
func GenerateEvtKey(namespace, kind, name, reason, message string) string {
	return fmt.Sprintf("%s/%s/%s/%s/%s", namespace, kind, name, reason, message)
}
