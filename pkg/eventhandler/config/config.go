package config

import (
	"github.com/weibaohui/k8m/pkg/models"
)

// EventHandlerConfig 定义事件处理器的完整配置
type EventHandlerConfig struct {
	Enabled      bool                    // 是否启用事件处理器
	Watcher      WatcherConfig           // Watcher配置
	Worker       WorkerConfig            // Worker配置
	EventConfigs []models.K8sEventConfig // 规则列表
	ClusterRules map[string]RuleConfig   // 集群级规则配置；key 为集群ID/名称，value 为该集群的事件过滤规则
	Webhooks     map[string][]string     // WebhookID列表，key 为集群ID/名称，value 为该集群的WebhookID列表
}

// WatcherConfig Watcher配置
type WatcherConfig struct {
	BufferSize int `json:"buffer_size" yaml:"buffer_size"` // 事件缓冲区大小
}

// WorkerConfig Worker配置
type WorkerConfig struct {
	BatchSize       int `json:"batch_size" yaml:"batch_size"`             // 批处理大小
	ProcessInterval int `json:"process_interval" yaml:"process_interval"` // 处理间隔(秒)
	MaxRetries      int `json:"max_retries" yaml:"max_retries"`           // 最大重试次数
}

// RuleConfig 定义事件过滤规则配置
type RuleConfig struct {
	Namespaces []string `json:"namespaces" yaml:"namespaces"` // 命名空间白名单/黑名单
	Names      []string `json:"names" yaml:"names"`           // 命名白名单/黑名单
	Reasons    []string `json:"reasons" yaml:"reasons"`       // 原因匹配
	Reverse    bool     `json:"reverse" yaml:"reverse"`       // 反向选择开关
}

// IsEmpty 判断规则配置是否为空
func (r *RuleConfig) IsEmpty() bool {
	return len(r.Namespaces) == 0 && len(r.Names) == 0 && len(r.Reasons) == 0
}

// DefaultEventHandlerConfig 创建默认的事件处理器配置
func DefaultEventHandlerConfig() *EventHandlerConfig {
	if cfg := LoadAllFromDB(); cfg != nil {
		return cfg
	}

	// 如果是nil，说明数据库中没有配置，返回默认配置，但是应是关闭的
	return &EventHandlerConfig{
		Enabled: false,
		Watcher: WatcherConfig{
			BufferSize: 1000,
		},
		Worker: WorkerConfig{
			BatchSize:       50,
			ProcessInterval: 1,
			MaxRetries:      3,
		},
		EventConfigs: []models.K8sEventConfig{},
	}
}
