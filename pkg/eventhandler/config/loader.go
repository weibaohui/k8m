// Package config 实现配置加载和初始化
package config

import (
	"fmt"
	"os"

	"github.com/weibaohui/k8m/pkg/eventhandler/model"
	"gopkg.in/yaml.v3"
	"k8s.io/klog/v2"
)

// LoadConfigFromFile 从文件加载配置
func LoadConfigFromFile(configPath string) (*model.EventHandlerConfig, error) {
	// 如果文件不存在，创建默认配置
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		klog.V(6).Infof("配置文件不存在，创建默认配置: %s", configPath)
		config := model.NewEventHandlerConfig()
		if err := SaveConfigToFile(configPath, config); err != nil {
			return nil, fmt.Errorf("保存默认配置失败: %w", err)
		}
		return config, nil
	}
	
	// 读取配置文件
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}
	
	// 解析配置
	var config model.EventHandlerConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}
	
	// 验证配置
	if err := ValidateConfig(&config); err != nil {
		return nil, fmt.Errorf("配置验证失败: %w", err)
	}
	
	klog.V(6).Infof("配置加载成功: %s", configPath)
	return &config, nil
}

// SaveConfigToFile 保存配置到文件
func SaveConfigToFile(configPath string, config *model.EventHandlerConfig) error {
	// 验证配置
	if err := ValidateConfig(config); err != nil {
		return fmt.Errorf("配置验证失败: %w", err)
	}
	
	// 序列化配置
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}
	
	// 写入文件
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}
	
	klog.V(6).Infof("配置保存成功: %s", configPath)
	return nil
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