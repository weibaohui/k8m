package webhook

import (
	"fmt"
	"sync"
)

// SendResult represents the result of a webhook send operation.
type SendResult struct {
	Status     string `json:"status"`
	StatusCode int    `json:"status_code"`
	RespBody   string `json:"resp_body"`
	Error      error  `json:"-"`
}

// PlatformAdapter defines the interface for platform-specific webhook adapters.
type PlatformAdapter interface {
	// Name returns the platform name (e.g., "dingtalk", "feishu", "wechat")
	Name() string

	// GetContentType returns the content type for HTTP requests
	GetContentType() string

	// FormatMessage formats the message according to platform requirements
	FormatMessage(msg, raw string, config *WebhookConfig) ([]byte, error)

	// SignRequest signs the request if required by the platform
	SignRequest(baseURL string, body []byte, secret string) (string, error)
}

// Global adapter registry
var (
	adapters     = make(map[string]PlatformAdapter)
	adapterMutex sync.RWMutex
)

// RegisterAdapter registers a platform adapter.
func RegisterAdapter(platform string, adapter PlatformAdapter) {
	adapterMutex.Lock()
	defer adapterMutex.Unlock()
	adapters[platform] = adapter
}

// GetAdapter retrieves a platform adapter by name.
func GetAdapter(platform string) (PlatformAdapter, error) {
	adapterMutex.RLock()
	defer adapterMutex.RUnlock()

	adapter, exists := adapters[platform]
	if !exists {
		return nil, fmt.Errorf("unknown platform: %s", platform)
	}
	return adapter, nil
}

// GetRegisteredPlatforms returns a list of all registered platform names.
func GetRegisteredPlatforms() []string {
	adapterMutex.RLock()
	defer adapterMutex.RUnlock()

	platforms := make([]string, 0, len(adapters))
	for platform := range adapters {
		platforms = append(platforms, platform)
	}
	return platforms
}
