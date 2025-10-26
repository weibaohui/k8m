package webhook

import (
	"testing"
	"time"

	"github.com/weibaohui/k8m/pkg/models"
)

func TestWebhookConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   *WebhookConfig
		wantErr  bool
	}{
		{
			name: "valid dingtalk config",
			config: &WebhookConfig{
				Platform:  "dingtalk",
				TargetURL: "https://oapi.dingtalk.com/robot/send?access_token=test",
			},
			wantErr: false,
		},
		{
			name: "valid feishu config",
			config: &WebhookConfig{
				Platform:  "feishu",
				TargetURL: "https://open.feishu.cn/open-apis/bot/v2/hook/test",
			},
			wantErr: false,
		},
		{
			name: "invalid platform",
			config: &WebhookConfig{
				Platform:  "invalid",
				TargetURL: "https://example.com",
			},
			wantErr: true,
		},
		{
			name: "invalid URL",
			config: &WebhookConfig{
				Platform:  "dingtalk",
				TargetURL: "not-a-url",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("WebhookConfig.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPlatformAdapters(t *testing.T) {
	// Ensure adapters are registered
	RegisterAllAdapters()

	platforms := []string{"dingtalk", "feishu", "wechat", "default"}
	
	for _, platform := range platforms {
		t.Run(platform, func(t *testing.T) {
			adapter, err := GetAdapter(platform)
			if err != nil {
				t.Fatalf("GetAdapter(%s) failed: %v", platform, err)
			}
			
			if adapter.Name() != platform {
				t.Errorf("Expected adapter name %s, got %s", platform, adapter.Name())
			}
			
			// Test message formatting
			config := &WebhookConfig{
				Platform:  platform,
				TargetURL: "https://example.com/webhook",
			}
			
			msg := "Test message"
			raw := "Raw test data"
			
			body, err := adapter.FormatMessage(msg, raw, config)
			if err != nil {
				t.Errorf("FormatMessage failed for %s: %v", platform, err)
			}
			
			if len(body) == 0 {
				t.Errorf("FormatMessage returned empty body for %s", platform)
			}
			
			// Test content type
			contentType := adapter.GetContentType()
			if contentType == "" {
				t.Errorf("GetContentType returned empty string for %s", platform)
			}
		})
	}
}

func TestWebhookClient(t *testing.T) {
	client := NewWebhookClient()
	if client == nil {
		t.Fatal("NewWebhookClient() returned nil")
	}
	
	clientWithTimeout := NewWebhookClientWithTimeout(10 * time.Second)
	if clientWithTimeout == nil {
		t.Fatal("NewWebhookClientWithTimeout() returned nil")
	}
}

func TestPushMsgToSingleTarget(t *testing.T) {
	// Test with a mock receiver
	receiver := &models.WebhookReceiver{
		Platform:  "default",
		TargetURL: "https://httpbin.org/post", // Using httpbin for testing
	}
	
	msg := "Test webhook message"
	raw := "Raw test data"
	
	// This will make an actual HTTP request to httpbin
	// In a real test environment, you might want to mock this
	result := PushMsgToSingleTarget(msg, raw, receiver)
	
	if result == nil {
		t.Fatal("PushMsgToSingleTarget returned nil result")
	}
	
	// For httpbin, we expect success
	if result.Status != "success" {
		t.Logf("Expected success status, got: %s (this might be expected if httpbin is not available)", result.Status)
	}
}

func TestBackwardCompatibility(t *testing.T) {
	// Test that RegisterAllSenders function exists and can be called (for API compatibility)
	RegisterAllSenders() // Should not panic
}