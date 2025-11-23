// Package config 实现配置加载和初始化
package config

import (
	"fmt"

	"github.com/weibaohui/k8m/pkg/eventhandler/model"
	"github.com/weibaohui/k8m/pkg/flag"
	"k8s.io/klog/v2"
)

// LoadConfigFromFlags 从flag配置加载事件处理器配置
func LoadConfigFromFlags() *model.EventHandlerConfig {
	cfg := flag.Init()

	// 创建基于flag的配置
	config := &model.EventHandlerConfig{
		Enabled: true,
		Database: model.DatabaseConfig{
			Type:     cfg.DBDriver,
			DSN:      getDSN(cfg),
			MaxConns: 10, // 使用默认值
		},
		Watcher: model.WatcherConfig{
			Enabled:        true,
			ResyncInterval: 300,
			BufferSize:     1000,
		},
		Worker: model.WorkerConfig{
			Enabled:         true,
			BatchSize:       100,
			ProcessInterval: 5000,
			MaxRetries:      3,
		},
		RuleConfig: *model.NewRuleConfig(),
		Webhook:    *model.NewWebhookConfig(),
	}

	// 如果数据库类型为空，使用默认值
	if config.Database.Type == "" {
		config.Database.Type = "sqlite"
	}

	klog.V(6).Infof("从flag配置加载事件处理器配置完成")
	return config
}

// getDSN 根据配置获取DSN
func getDSN(cfg *flag.Config) string {
	switch cfg.DBDriver {
	case "sqlite":
		return cfg.SqliteDSN
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&collation=%s&%s",
			cfg.MysqlUser, cfg.MysqlPassword, cfg.MysqlHost, cfg.MysqlPort,
			cfg.MysqlDatabase, cfg.MysqlCharset, cfg.MysqlCollation, cfg.MysqlQuery)
	case "postgresql":
		return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=%s",
			cfg.PgHost, cfg.PgUser, cfg.PgPassword, cfg.PgDatabase,
			cfg.PgPort, cfg.PgSSLMode, cfg.PgTimeZone)
	default:
		return cfg.SqliteDSN
	}
}

// ValidateConfig 验证配置
func ValidateConfig(config *model.EventHandlerConfig) error {
	if config == nil {
		return fmt.Errorf("配置不能为空")
	}

	// 验证数据库配置
	if err := ValidateDatabaseConfig(&config.Database); err != nil {
		return fmt.Errorf("数据库配置验证失败: %w", err)
	}

	// 验证Watcher配置
	if err := ValidateWatcherConfig(&config.Watcher); err != nil {
		return fmt.Errorf("Watcher配置验证失败: %w", err)
	}

	// 验证Worker配置
	if err := ValidateWorkerConfig(&config.Worker); err != nil {
		return fmt.Errorf("Worker配置验证失败: %w", err)
	}

	// 验证Webhook配置
	if err := ValidateWebhookConfig(&config.Webhook); err != nil {
		return fmt.Errorf("Webhook配置验证失败: %w", err)
	}

	return nil
}

// ValidateDatabaseConfig 验证数据库配置
func ValidateDatabaseConfig(config *model.DatabaseConfig) error {
	if config.Type == "" {
		config.Type = "sqlite"
	}

	validTypes := []string{"sqlite", "postgres", "mysql"}
	isValid := false
	for _, t := range validTypes {
		if config.Type == t {
			isValid = true
			break
		}
	}

	if !isValid {
		return fmt.Errorf("不支持的数据库类型: %s", config.Type)
	}

	if config.DSN == "" {
		switch config.Type {
		case "sqlite":
			config.DSN = "k8s_events.db"
		case "postgres":
			config.DSN = "host=localhost user=postgres password=postgres dbname=k8s_events sslmode=disable"
		case "mysql":
			config.DSN = "user:password@tcp(localhost:3306)/k8s_events?charset=utf8mb4&parseTime=True&loc=Local"
		}
	}

	if config.MaxConns <= 0 {
		config.MaxConns = 10
	}

	return nil
}

// ValidateWatcherConfig 验证Watcher配置
func ValidateWatcherConfig(config *model.WatcherConfig) error {
	if config.ResyncInterval <= 0 {
		config.ResyncInterval = 300
	}

	if config.BufferSize <= 0 {
		config.BufferSize = 1000
	}

	return nil
}

// ValidateWorkerConfig 验证Worker配置
func ValidateWorkerConfig(config *model.WorkerConfig) error {
	if config.BatchSize <= 0 {
		config.BatchSize = 100
	}

	if config.ProcessInterval <= 0 {
		config.ProcessInterval = 5000
	}

	if config.MaxRetries <= 0 {
		config.MaxRetries = 3
	}

	return nil
}

// ValidateWebhookConfig 验证Webhook配置
func ValidateWebhookConfig(config *model.WebhookConfig) error {
	if config.Enabled && config.URL == "" {
		return fmt.Errorf("Webhook启用时URL不能为空")
	}

	if config.Method == "" {
		config.Method = "POST"
	}

	if config.Timeout <= 0 {
		config.Timeout = 30
	}

	if config.Retries <= 0 {
		config.Retries = 3
	}

	return nil
}
