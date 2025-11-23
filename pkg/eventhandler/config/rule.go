package config

// RuleConfig 定义事件过滤规则配置
type RuleConfig struct {
	Namespaces []string          `json:"namespaces" yaml:"namespaces"` // 命名空间白名单/黑名单
	Labels     map[string]string `json:"labels" yaml:"labels"`         // 标签匹配
	Reasons    []string          `json:"reasons" yaml:"reasons"`       // 原因匹配
	Types      []string          `json:"types" yaml:"types"`           // 事件类型匹配
	Reverse    bool              `json:"reverse" yaml:"reverse"`       // 反向选择开关
}

// NewRuleConfig 创建新的规则配置
func NewRuleConfig() *RuleConfig {
	return &RuleConfig{
		Namespaces: []string{},
		Labels:     make(map[string]string),
		Reasons:    []string{},
		Types:      []string{},
		Reverse:    false,
	}
}

// IsEmpty 判断规则配置是否为空
func (r *RuleConfig) IsEmpty() bool {
	return len(r.Namespaces) == 0 && len(r.Labels) == 0 && len(r.Reasons) == 0 && len(r.Types) == 0
}
