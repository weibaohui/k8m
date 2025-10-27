package webhook

import (
	"net/url"

	"github.com/weibaohui/k8m/pkg/models"
)

// WebhookConfig represents the configuration for a webhook endpoint.
// This replaces the old Channel concept with clearer naming and responsibilities.
type WebhookConfig struct {
	WebhookId    uint   // WebhookID of the webhook configuration
	WebhookName  string // WebhookName of the webhook configuration
	Platform     string // Platform identifier (feishu, dingtalk, wechat, default)
	TargetURL    string // The webhook endpoint URL
	BodyTemplate string // Message body template (optional, platform defaults will be used if empty)
	SignSecret   string // Secret for signing requests (platform-specific)
}

// NewWebhookConfig creates a new webhook configuration from a WebhookReceiver model.
func NewWebhookConfig(receiver *models.WebhookReceiver) *WebhookConfig {
	return &WebhookConfig{
		WebhookId:    receiver.ID,
		WebhookName:  receiver.Name,
		Platform:     receiver.Platform,
		TargetURL:    receiver.TargetURL,
		BodyTemplate: receiver.BodyTemplate,
		SignSecret:   receiver.SignSecret,
	}
}

// GetDefaultTemplate returns the default message template for the platform.
func (c *WebhookConfig) GetDefaultTemplate() string {
	switch c.Platform {
	case "feishu":
		return `{"msg_type":"text","content":{"text":"%s"}}`
	case "dingtalk":
		return `{"msgtype":"text","text":{"content":"%s"}}`
	case "wechat":
		return `{"msgtype":"markdown","markdown":{"content":"%s"}}`
	default:
		return "%s" // Simple text format for custom platforms
	}
}

// GetEffectiveTemplate returns the template to use, falling back to default if none specified.
func (c *WebhookConfig) GetEffectiveTemplate() string {
	if c.BodyTemplate != "" {
		return c.BodyTemplate
	}
	return c.GetDefaultTemplate()
}

// HasSignature returns true if this configuration requires request signing.
func (c *WebhookConfig) HasSignature() bool {
	return c.SignSecret != ""
}

// Validate checks if the configuration is valid.
func (c *WebhookConfig) Validate() error {
	if c.Platform == "" {
		return ErrInvalidPlatform
	}

	// Validate platform is supported
	validPlatforms := map[string]bool{
		"dingtalk": true,
		"feishu":   true,
		"wechat":   true,
		"default":  true,
	}
	if !validPlatforms[c.Platform] {
		return ErrInvalidPlatform
	}

	if c.TargetURL == "" {
		return ErrInvalidURL
	}

	// Validate URL format
	parsedURL, err := url.Parse(c.TargetURL)
	if err != nil {
		return ErrInvalidURL
	}

	// Ensure it's a valid HTTP/HTTPS URL
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return ErrInvalidURL
	}

	if parsedURL.Host == "" {
		return ErrInvalidURL
	}

	return nil
}
