// Package model 定义事件处理器的配置模型
package model

// EventHandlerConfig 定义事件处理器的完整配置
type EventHandlerConfig struct {
	Enabled      bool           `json:"enabled" yaml:"enabled"`           // 是否启用事件处理器
	Database     DatabaseConfig `json:"database" yaml:"database"`         // 数据库配置
	Watcher      WatcherConfig  `json:"watcher" yaml:"watcher"`           // Watcher配置
	Worker       WorkerConfig   `json:"worker" yaml:"worker"`             // Worker配置
	RuleConfig   RuleConfig     `json:"rule_config" yaml:"rule_config"`   // 规则配置
	Webhook      WebhookConfig  `json:"webhook" yaml:"webhook"`           // Webhook配置
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type     string `json:"type" yaml:"type"`         // 数据库类型: sqlite, postgres, mysql
	DSN      string `json:"dsn" yaml:"dsn"`           // 数据库连接字符串
	MaxConns int    `json:"max_conns" yaml:"max_conns"` // 最大连接数
}

// WatcherConfig Watcher配置
type WatcherConfig struct {
	Enabled         bool   `json:"enabled" yaml:"enabled"`                   // 是否启用Watcher
	ResyncInterval  int    `json:"resync_interval" yaml:"resync_interval"`   // 重新同步间隔(秒)
	BufferSize      int    `json:"buffer_size" yaml:"buffer_size"`           // 事件缓冲区大小
}

// WorkerConfig Worker配置
type WorkerConfig struct {
	Enabled        bool `json:"enabled" yaml:"enabled"`                 // 是否启用Worker
	BatchSize      int  `json:"batch_size" yaml:"batch_size"`             // 批处理大小
	ProcessInterval int `json:"process_interval" yaml:"process_interval"` // 处理间隔(毫秒)
	MaxRetries     int  `json:"max_retries" yaml:"max_retries"`             // 最大重试次数
}

// NewEventHandlerConfig 创建默认的事件处理器配置
func NewEventHandlerConfig() *EventHandlerConfig {
	return &EventHandlerConfig{
		Enabled: true,
		Database: DatabaseConfig{
			Type:     "sqlite",
			DSN:      "k8s_events.db",
			MaxConns: 10,
		},
		Watcher: WatcherConfig{
			Enabled:        true,
			ResyncInterval: 300,
			BufferSize:     1000,
		},
		Worker: WorkerConfig{
			Enabled:         true,
			BatchSize:       50,
			ProcessInterval: 500,
			MaxRetries:      3,
		},
		RuleConfig: *NewRuleConfig(),
		Webhook:    *NewWebhookConfig(),
	}
}