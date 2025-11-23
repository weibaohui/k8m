// Package config 实现配置管理
package config

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/weibaohui/k8m/pkg/eventhandler/model"
	"github.com/weibaohui/k8m/pkg/eventhandler/watcher"
	"github.com/weibaohui/k8m/pkg/eventhandler/worker"
	"gopkg.in/yaml.v3"
	"k8s.io/klog/v2"
)

// ConfigManager 配置管理器
type ConfigManager struct {
	config      *model.EventHandlerConfig
	mu          sync.RWMutex
	watcher     *watcher.EventWatcher
	worker      *worker.EventWorker
	configPath  string
	lastModTime time.Time
}

// NewConfigManager 创建配置管理器
func NewConfigManager(configPath string, watcher *watcher.EventWatcher, worker *worker.EventWorker) *ConfigManager {
	return &ConfigManager{
		configPath: configPath,
		watcher:    watcher,
		worker:     worker,
	}
}

// LoadConfig 加载配置文件
func (c *ConfigManager) LoadConfig() error {
	data, err := os.ReadFile(c.configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config model.EventHandlerConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	c.mu.Lock()
	c.config = &config
	c.mu.Unlock()

	// 更新文件修改时间
	if info, err := os.Stat(c.configPath); err == nil {
		c.lastModTime = info.ModTime()
	}

	klog.V(6).Infof("配置加载成功: %s", c.configPath)
	return nil
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

// StartAutoReload 启动自动重载
func (c *ConfigManager) StartAutoReload(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := c.checkAndReload(); err != nil {
				klog.Errorf("配置重载检查失败: %v", err)
			}
		}
	}
}

// checkAndReload 检查并重新加载配置
func (c *ConfigManager) checkAndReload() error {
	info, err := os.Stat(c.configPath)
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %w", err)
	}

	// 检查文件是否被修改
	if info.ModTime().After(c.lastModTime) {
		klog.V(6).Infof("检测到配置文件变更，开始重载")
		return c.LoadConfig()
	}

	return nil
}

// SaveConfig 保存配置到文件
func (c *ConfigManager) SaveConfig() error {
	c.mu.RLock()
	data, err := yaml.Marshal(c.config)
	c.mu.RUnlock()

	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(c.configPath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	klog.V(6).Infof("配置保存成功: %s", c.configPath)
	return nil
}
