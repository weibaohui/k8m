// Package store 实现事件存储
package store

import (
	"context"
	"fmt"
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/eventhandler/model"
	"github.com/weibaohui/k8m/pkg/models"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

// GORMStore GORM存储实现
type GORMStore struct {
	db *gorm.DB
}

// NewGORMStore 创建GORM存储实例
func NewGORMStore() *GORMStore {
	db := dao.DB()
	if db == nil {
		klog.Errorf("数据库连接失败")
		return nil
	}
	return &GORMStore{db: db}
}

// Init 初始化数据库表
func (s *GORMStore) Init() error {
	if s.db == nil {
		return fmt.Errorf("数据库连接为空")
	}

	err := s.db.AutoMigrate(&models.K8sEvent{})
	if err != nil {
		klog.Errorf("创建事件表失败: %v", err)
		return fmt.Errorf("创建事件表失败: %w", err)
	}

	klog.V(6).Infof("事件表初始化成功")
	return nil
}

// Create 创建新事件
func (s *GORMStore) Create(ctx context.Context, event *model.Event) error {
	if s.db == nil {
		return fmt.Errorf("数据库连接为空")
	}

	k8sEvent := &models.K8sEvent{
		EvtKey:    event.EvtKey,
		Type:      event.Type,
		Reason:    event.Reason,
		Level:     event.Level,
		Namespace: event.Namespace,
		Name:      event.Name,
		Message:   event.Message,
		Timestamp: event.Timestamp,
		Processed: event.Processed,
		Attempts:  event.Attempts,
	}

	// 使用FirstOrCreate避免重复插入，如果存在则更新
	result := s.db.WithContext(ctx).Where("evt_key = ?", k8sEvent.EvtKey).FirstOrCreate(k8sEvent)
	if result.Error != nil {
		klog.Errorf("创建事件失败: %v", result.Error)
		return fmt.Errorf("创建事件失败: %w", result.Error)
	}

	// 如果记录已存在，更新timestamp和message
	if result.RowsAffected == 0 {
		updateResult := s.db.WithContext(ctx).Model(&models.K8sEvent{}).
			Where("evt_key = ?", k8sEvent.EvtKey).
			Updates(map[string]interface{}{
				"timestamp": k8sEvent.Timestamp,
				"message":   k8sEvent.Message,
			})
		if updateResult.Error != nil {
			klog.Errorf("更新事件失败: %v", updateResult.Error)
			return fmt.Errorf("更新事件失败: %w", updateResult.Error)
		}
	}

	klog.V(6).Infof("事件创建成功: %s", event.EvtKey)
	return nil
}

// GetByKey 根据事件键获取事件
func (s *GORMStore) GetByKey(ctx context.Context, evtKey string) (*model.Event, error) {
	if s.db == nil {
		return nil, fmt.Errorf("数据库连接为空")
	}

	var k8sEvent models.K8sEvent
	err := s.db.WithContext(ctx).Where("evt_key = ?", evtKey).First(&k8sEvent).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		klog.Errorf("查询事件失败: %v", err)
		return nil, fmt.Errorf("查询事件失败: %w", err)
	}

	return convertToEvent(&k8sEvent), nil
}

// GetUnprocessed 获取未处理的事件
func (s *GORMStore) GetUnprocessed(ctx context.Context, limit int) ([]*model.Event, error) {
	if s.db == nil {
		return nil, fmt.Errorf("数据库连接为空")
	}

	var k8sEvents []models.K8sEvent
	err := s.db.WithContext(ctx).
		Where("processed = ?", false).
		Order("timestamp ASC").
		Limit(limit).
		Find(&k8sEvents).Error

	if err != nil {
		klog.Errorf("查询未处理事件失败: %v", err)
		return nil, fmt.Errorf("查询未处理事件失败: %w", err)
	}

	events := make([]*model.Event, len(k8sEvents))
	for i, k8sEvent := range k8sEvents {
		events[i] = convertToEvent(&k8sEvent)
	}

	return events, nil
}

// UpdateProcessed 更新事件处理状态
func (s *GORMStore) UpdateProcessed(ctx context.Context, id int64, processed bool) error {
	if s.db == nil {
		return fmt.Errorf("数据库连接为空")
	}

	err := s.db.WithContext(ctx).Model(&models.K8sEvent{}).
		Where("id = ?", id).
		Update("processed", processed).Error

	if err != nil {
		klog.Errorf("更新事件处理状态失败: %v", err)
		return fmt.Errorf("更新事件处理状态失败: %w", err)
	}

	klog.V(6).Infof("事件处理状态更新成功: id=%d, processed=%t", id, processed)
	return nil
}

// IncrementAttempts 增加重试次数
func (s *GORMStore) IncrementAttempts(ctx context.Context, id int64) error {
	if s.db == nil {
		return fmt.Errorf("数据库连接为空")
	}

	err := s.db.WithContext(ctx).Model(&models.K8sEvent{}).
		Where("id = ?", id).
		UpdateColumn("attempts", gorm.Expr("attempts + ?", 1)).Error

	if err != nil {
		klog.Errorf("增加重试次数失败: %v", err)
		return fmt.Errorf("增加重试次数失败: %w", err)
	}

	return nil
}

// DeleteOldEvents 删除旧事件
func (s *GORMStore) DeleteOldEvents(ctx context.Context, days int) error {
	if s.db == nil {
		return fmt.Errorf("数据库连接为空")
	}

	// 计算删除时间
	deleteTime := time.Now().AddDate(0, 0, -days)

	result := s.db.WithContext(ctx).
		Where("timestamp < ?", deleteTime).
		Delete(&models.K8sEvent{})

	if result.Error != nil {
		klog.Errorf("删除旧事件失败: %v", result.Error)
		return fmt.Errorf("删除旧事件失败: %w", result.Error)
	}

	klog.V(6).Infof("删除旧事件成功: 删除%d条记录", result.RowsAffected)
	return nil
}

// Close 关闭存储连接
func (s *GORMStore) Close() error {
	// GORM会自动管理连接，这里不需要特别处理
	return nil
}

// convertToEvent 将K8sEvent转换为Event
func convertToEvent(k8sEvent *models.K8sEvent) *model.Event {
	return &model.Event{
		ID:        k8sEvent.ID,
		EvtKey:    k8sEvent.EvtKey,
		Type:      k8sEvent.Type,
		Reason:    k8sEvent.Reason,
		Level:     k8sEvent.Level,
		Namespace: k8sEvent.Namespace,
		Name:      k8sEvent.Name,
		Message:   k8sEvent.Message,
		Timestamp: k8sEvent.Timestamp,
		Processed: k8sEvent.Processed,
		Attempts:  k8sEvent.Attempts,
		CreatedAt: k8sEvent.CreatedAt,
		UpdatedAt: k8sEvent.UpdatedAt,
	}
}
