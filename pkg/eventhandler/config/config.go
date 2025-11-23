package config

// EventHandlerConfig 定义事件处理器的完整配置
type EventHandlerConfig struct {
	Enabled    bool          `json:"enabled" yaml:"enabled"`         // 是否启用事件处理器
	Watcher    WatcherConfig `json:"watcher" yaml:"watcher"`         // Watcher配置
	Worker     WorkerConfig  `json:"worker" yaml:"worker"`           // Worker配置
	RuleConfig RuleConfig    `json:"rule_config" yaml:"rule_config"` // 规则配置
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
	Namespaces []string          `json:"namespaces" yaml:"namespaces"` // 命名空间白名单/黑名单
	Labels     map[string]string `json:"labels" yaml:"labels"`         // 标签匹配
	Reasons    []string          `json:"reasons" yaml:"reasons"`       // 原因匹配
	Types      []string          `json:"types" yaml:"types"`           // 事件类型匹配
	Reverse    bool              `json:"reverse" yaml:"reverse"`       // 反向选择开关
}

// IsEmpty 判断规则配置是否为空
func (r *RuleConfig) IsEmpty() bool {
	return len(r.Namespaces) == 0 && len(r.Labels) == 0 && len(r.Reasons) == 0 && len(r.Types) == 0
}

// DefaultEventHandlerConfig 创建默认的事件处理器配置
func DefaultEventHandlerConfig() *EventHandlerConfig {
	// TODO 从配置界面中、数据库中加载配置选项
	if cfg := LoadFromDB(); cfg != nil {
		return cfg
	}
	return &EventHandlerConfig{
		Enabled: true,
		Watcher: WatcherConfig{
			BufferSize: 1000,
		},
		Worker: WorkerConfig{
			BatchSize:       50,
			ProcessInterval: 1,
			MaxRetries:      3,
		},
		RuleConfig: RuleConfig{
			Namespaces: []string{},
			Labels:     make(map[string]string),
			Reasons:    []string{},
			Types:      []string{},
			Reverse:    false,
		},
	}
}
