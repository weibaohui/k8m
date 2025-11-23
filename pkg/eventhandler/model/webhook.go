// Package model 定义Webhook配置相关的数据模型
package model

// WebhookConfig 定义Webhook配置
type WebhookConfig struct {
	Enabled bool              `json:"enabled" yaml:"enabled"` // 是否启用
	URL     string            `json:"url" yaml:"url"`         // Webhook URL
	Method  string            `json:"method" yaml:"method"`   // HTTP方法
	Headers map[string]string `json:"headers" yaml:"headers"` // 请求头
	Timeout int               `json:"timeout" yaml:"timeout"` // 超时时间(秒)
	Retries int               `json:"retries" yaml:"retries"` // 重试次数
}

// NewWebhookConfig 创建新的Webhook配置
func NewWebhookConfig() *WebhookConfig {
	return &WebhookConfig{
		Enabled: false,
		Method:  "POST",
		Headers: make(map[string]string),
		Timeout: 30,
		Retries: 3,
	}
}
