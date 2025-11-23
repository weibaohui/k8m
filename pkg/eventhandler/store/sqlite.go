// Package store 实现SQLite存储
package store

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/weibaohui/k8m/pkg/eventhandler/model"
	"k8s.io/klog/v2"
)

// SQLiteStore SQLite存储实现
type SQLiteStore struct {
	db      *sql.DB
	options *EventStoreOptions
}

// NewSQLiteStore 创建SQLite存储实例
func NewSQLiteStore(db *sql.DB, options *EventStoreOptions) *SQLiteStore {
	if options == nil {
		options = NewEventStoreOptions()
	}
	return &SQLiteStore{
		db:      db,
		options: options,
	}
}

// Init 初始化数据库表
func (s *SQLiteStore) Init() error {
	_, err := s.db.Exec(model.SQLiteEventTableSQL)
	if err != nil {
		klog.Errorf("创建事件表失败: %v", err)
		return fmt.Errorf("创建事件表失败: %w", err)
	}
	klog.V(6).Infof("事件表初始化成功")
	return nil
}

// Create 创建新事件
func (s *SQLiteStore) Create(ctx context.Context, event *model.Event) error {
	query := `
		INSERT INTO k8s_events (evt_key, type, reason, level, namespace, name, message, timestamp, processed, attempts)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(evt_key) DO UPDATE SET
			timestamp = excluded.timestamp,
			message = excluded.message,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err := s.db.ExecContext(ctx, query,
		event.EvtKey,
		event.Type,
		event.Reason,
		event.Level,
		event.Namespace,
		event.Name,
		event.Message,
		event.Timestamp,
		event.Processed,
		event.Attempts,
	)

	if err != nil {
		klog.Errorf("创建事件失败: %v", err)
		return fmt.Errorf("创建事件失败: %w", err)
	}

	klog.V(6).Infof("事件创建成功: %s", event.EvtKey)
	return nil
}

// GetByKey 根据事件键获取事件
func (s *SQLiteStore) GetByKey(ctx context.Context, evtKey string) (*model.Event, error) {
	query := `
		SELECT id, evt_key, type, reason, level, namespace, name, message, timestamp, processed, attempts, created_at, updated_at
		FROM k8s_events
		WHERE evt_key = ?
	`

	var event model.Event
	err := s.db.QueryRowContext(ctx, query, evtKey).Scan(
		&event.ID, &event.EvtKey, &event.Type, &event.Reason, &event.Level,
		&event.Namespace, &event.Name, &event.Message, &event.Timestamp,
		&event.Processed, &event.Attempts, &event.CreatedAt, &event.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		klog.Errorf("查询事件失败: %v", err)
		return nil, fmt.Errorf("查询事件失败: %w", err)
	}

	return &event, nil
}

// GetUnprocessed 获取未处理的事件
func (s *SQLiteStore) GetUnprocessed(ctx context.Context, limit int) ([]*model.Event, error) {
	query := `
		SELECT id, evt_key, type, reason, level, namespace, name, message, timestamp, processed, attempts, created_at, updated_at
		FROM k8s_events
		WHERE processed = false
		ORDER BY timestamp ASC
		LIMIT ?
	`

	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		klog.Errorf("查询未处理事件失败: %v", err)
		return nil, fmt.Errorf("查询未处理事件失败: %w", err)
	}
	defer rows.Close()

	var events []*model.Event
	for rows.Next() {
		var event model.Event
		err := rows.Scan(
			&event.ID, &event.EvtKey, &event.Type, &event.Reason, &event.Level,
			&event.Namespace, &event.Name, &event.Message, &event.Timestamp,
			&event.Processed, &event.Attempts, &event.CreatedAt, &event.UpdatedAt,
		)
		if err != nil {
			klog.Errorf("扫描事件失败: %v", err)
			continue
		}
		events = append(events, &event)
	}

	return events, nil
}

// UpdateProcessed 更新事件处理状态
func (s *SQLiteStore) UpdateProcessed(ctx context.Context, id int64, processed bool) error {
	query := `
		UPDATE k8s_events
		SET processed = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := s.db.ExecContext(ctx, query, processed, id)
	if err != nil {
		klog.Errorf("更新事件处理状态失败: %v", err)
		return fmt.Errorf("更新事件处理状态失败: %w", err)
	}

	klog.V(6).Infof("事件处理状态更新成功: id=%d, processed=%t", id, processed)
	return nil
}

// IncrementAttempts 增加重试次数
func (s *SQLiteStore) IncrementAttempts(ctx context.Context, id int64) error {
	query := `
		UPDATE k8s_events
		SET attempts = attempts + 1, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		klog.Errorf("增加重试次数失败: %v", err)
		return fmt.Errorf("增加重试次数失败: %w", err)
	}

	return nil
}

// DeleteOldEvents 删除旧事件
func (s *SQLiteStore) DeleteOldEvents(ctx context.Context, days int) error {
	query := `
		DELETE FROM k8s_events
		WHERE timestamp < datetime('now', '-' || ? || ' days')
	`

	result, err := s.db.ExecContext(ctx, query, days)
	if err != nil {
		klog.Errorf("删除旧事件失败: %v", err)
		return fmt.Errorf("删除旧事件失败: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	klog.V(6).Infof("删除旧事件成功: 删除%d条记录", rowsAffected)
	return nil
}

// Close 关闭存储连接
func (s *SQLiteStore) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}
