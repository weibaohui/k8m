// Package config 实现配置加载和初始化
package config

import (
	"fmt"

	"github.com/weibaohui/k8m/pkg/eventhandler/model"
	"k8s.io/klog/v2"
)

// LoadConfigFromFlags 从flag配置加载事件处理器配置
func LoadConfigFromFlags() *model.EventHandlerConfig {

	// 创建基于flag的配置
	config := &model.EventHandlerConfig{
		Enabled: true,

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
	}

	klog.V(6).Infof("从flag配置加载事件处理器配置完成")
	return config
}

// ValidateConfig 验证配置
func ValidateConfig(config *model.EventHandlerConfig) error {
	if config == nil {
		return fmt.Errorf("配置不能为空")
	}

	// 验证Watcher配置
	if err := ValidateWatcherConfig(&config.Watcher); err != nil {
		return fmt.Errorf("Watcher配置验证失败: %w", err)
	}

	// 验证Worker配置
	if err := ValidateWorkerConfig(&config.Worker); err != nil {
		return fmt.Errorf("Worker配置验证失败: %w", err)
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
