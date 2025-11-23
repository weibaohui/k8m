// Package store 定义事件存储接口和实现
package store

import (
	"context"
	"time"

	"github.com/weibaohui/k8m/pkg/eventhandler/model"
)

// EventStore 定义事件存储接口
type EventStore interface {
	// Create 创建新事件
	Create(ctx context.Context, event *model.Event) error
	
	// GetByKey 根据事件键获取事件
	GetByKey(ctx context.Context, evtKey string) (*model.Event, error)
	
	// GetUnprocessed 获取未处理的事件
	GetUnprocessed(ctx context.Context, limit int) ([]*model.Event, error)
	
	// UpdateProcessed 更新事件处理状态
	UpdateProcessed(ctx context.Context, id int64, processed bool) error
	
	// IncrementAttempts 增加重试次数
	IncrementAttempts(ctx context.Context, id int64) error
	
	// DeleteOldEvents 删除旧事件
	DeleteOldEvents(ctx context.Context, days int) error
	
	// Close 关闭存储连接
	Close() error
}

// EventStoreOptions 存储配置选项
type EventStoreOptions struct {
	MaxRetries int
	Timeout    time.Duration
}

// NewEventStoreOptions 创建默认存储配置
func NewEventStoreOptions() *EventStoreOptions {
	return &EventStoreOptions{
		MaxRetries: 3,
		Timeout:    30 * time.Second,
	}
}