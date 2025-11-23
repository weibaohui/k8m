// Package model 定义数据库表结构
package model

// EventTableSQL 创建事件表的SQL语句
const EventTableSQL = `
CREATE TABLE IF NOT EXISTS k8s_events (
    id BIGSERIAL PRIMARY KEY,
    evt_key TEXT UNIQUE NOT NULL,
    type VARCHAR(16) NOT NULL,
    reason VARCHAR(128) NOT NULL,
    level VARCHAR(16) NOT NULL,
    namespace VARCHAR(64) NOT NULL,
    name VARCHAR(128) NOT NULL,
    message TEXT NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    processed BOOLEAN DEFAULT FALSE,
    attempts INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_k8s_events_processed_timestamp ON k8s_events(processed, timestamp);
CREATE INDEX IF NOT EXISTS idx_k8s_events_namespace ON k8s_events(namespace);
CREATE INDEX IF NOT EXISTS idx_k8s_events_type ON k8s_events(type);
CREATE INDEX IF NOT EXISTS idx_k8s_events_reason ON k8s_events(reason);
CREATE INDEX IF NOT EXISTS idx_k8s_events_evt_key ON k8s_events(evt_key);
`

// SQLiteEventTableSQL SQLite版本的创建事件表SQL语句
const SQLiteEventTableSQL = `
CREATE TABLE IF NOT EXISTS k8s_events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    evt_key TEXT UNIQUE NOT NULL,
    type TEXT NOT NULL,
    reason TEXT NOT NULL,
    level TEXT NOT NULL,
    namespace TEXT NOT NULL,
    name TEXT NOT NULL,
    message TEXT NOT NULL,
    timestamp DATETIME NOT NULL,
    processed BOOLEAN DEFAULT FALSE,
    attempts INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_k8s_events_processed_timestamp ON k8s_events(processed, timestamp);
CREATE INDEX IF NOT EXISTS idx_k8s_events_namespace ON k8s_events(namespace);
CREATE INDEX IF NOT EXISTS idx_k8s_events_type ON k8s_events(type);
CREATE INDEX IF NOT EXISTS idx_k8s_events_reason ON k8s_events(reason);
CREATE INDEX IF NOT EXISTS idx_k8s_events_evt_key ON k8s_events(evt_key);
`

// MySQLEventTableSQL MySQL版本的创建事件表SQL语句
const MySQLEventTableSQL = `
CREATE TABLE IF NOT EXISTS k8s_events (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    evt_key VARCHAR(255) UNIQUE NOT NULL,
    type VARCHAR(16) NOT NULL,
    reason VARCHAR(128) NOT NULL,
    level VARCHAR(16) NOT NULL,
    namespace VARCHAR(64) NOT NULL,
    name VARCHAR(128) NOT NULL,
    message TEXT NOT NULL,
    timestamp TIMESTAMP NOT NULL,
    processed BOOLEAN DEFAULT FALSE,
    attempts INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_processed_timestamp (processed, timestamp),
    INDEX idx_namespace (namespace),
    INDEX idx_type (type),
    INDEX idx_reason (reason),
    INDEX idx_evt_key (evt_key)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
`
