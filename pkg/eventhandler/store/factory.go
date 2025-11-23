// Package store 存储工厂函数
package store

import (
	"database/sql"
	"fmt"

	"github.com/weibaohui/k8m/pkg/eventhandler/model"
	"k8s.io/klog/v2"
)

// NewStore 根据配置创建存储实例
func NewStore(dbType string, db *sql.DB, options *EventStoreOptions) (EventStore, error) {
	switch dbType {
	case "sqlite":
		sqliteStore := NewSQLiteStore(db, options)
		if err := sqliteStore.Init(); err != nil {
			return nil, err
		}
		return sqliteStore, nil
	case "postgres", "mysql":
		// TODO: 实现PostgreSQL和MySQL存储
		return nil, fmt.Errorf("数据库类型 %s 暂未实现", dbType)
	default:
		return nil, fmt.Errorf("不支持的数据库类型: %s", dbType)
	}
}

// CreateTable 根据数据库类型创建表
func CreateTable(dbType string, db *sql.DB) error {
	switch dbType {
	case "sqlite":
		_, err := db.Exec(model.SQLiteEventTableSQL)
		if err != nil {
			return fmt.Errorf("创建SQLite事件表失败: %w", err)
		}
	case "postgres":
		_, err := db.Exec(model.EventTableSQL)
		if err != nil {
			return fmt.Errorf("创建PostgreSQL事件表失败: %w", err)
		}
	case "mysql":
		_, err := db.Exec(model.MySQLEventTableSQL)
		if err != nil {
			return fmt.Errorf("创建MySQL事件表失败: %w", err)
		}
	default:
		return fmt.Errorf("不支持的数据库类型: %s", dbType)
	}
	
	klog.V(6).Infof("事件表创建成功，数据库类型: %s", dbType)
	return nil
}