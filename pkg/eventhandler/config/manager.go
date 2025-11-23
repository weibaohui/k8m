// Package config 实现配置管理
package config

import (
	"sync"

	"github.com/weibaohui/k8m/pkg/eventhandler/model"
	"github.com/weibaohui/k8m/pkg/eventhandler/watcher"
	"github.com/weibaohui/k8m/pkg/eventhandler/worker"
	"k8s.io/klog/v2"
)

// ConfigManager 配置管理器
type ConfigManager struct {
	config  *model.EventHandlerConfig
	mu      sync.RWMutex
	watcher *watcher.EventWatcher
	worker  *worker.EventWorker
}

// NewConfigManager 创建配置管理器
func NewConfigManager(watcher *watcher.EventWatcher, worker *worker.EventWorker) *ConfigManager {
	config := LoadConfigFromFlags()
	return &ConfigManager{
		config:  config,
		watcher: watcher,
		worker:  worker,
	}
}

// GetConfig 获取当前配置
func (c *ConfigManager) GetConfig() *model.EventHandlerConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config
}

// UpdateRuleConfig 更新规则配置
func (c *ConfigManager) UpdateRuleConfig(ruleConfig *model.RuleConfig) error {
	c.mu.Lock()
	if c.config == nil {
		c.config = &model.EventHandlerConfig{}
	}
	c.config.RuleConfig = *ruleConfig
	c.mu.Unlock()

	// 更新Watcher的规则匹配器
	if c.watcher != nil {
		// TODO: 实现Watcher的规则更新方法
		klog.V(6).Infof("规则配置更新成功，需要重启Watcher生效")
	}

	klog.V(6).Infof("规则配置更新成功")
	return nil
}
