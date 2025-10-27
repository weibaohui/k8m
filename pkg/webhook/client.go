package webhook

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"k8s.io/klog/v2"
)

// WebhookClient provides a unified HTTP transport layer for webhook sending.
type WebhookClient struct {
	timeout time.Duration
}

// NewWebhookClient creates a new webhook client with default settings.
func NewWebhookClient() *WebhookClient {
	return &WebhookClient{
		timeout: 30 * time.Second,
	}
}

// NewWebhookClientWithTimeout creates a new webhook client with custom timeout.
func NewWebhookClientWithTimeout(timeout time.Duration) *WebhookClient {
	return &WebhookClient{
		timeout: timeout,
	}
}

// Send sends a webhook message using the specified configuration and platform adapter.
func (c *WebhookClient) Send(ctx context.Context, msg, raw string, config *WebhookConfig) (*SendResult, error) {
	// Validate configuration
	if err := config.Validate(); err != nil {
		return &SendResult{
			Status:   "failed",
			RespBody: err.Error(),
			Error:    err,
		}, err
	}

	// Get platform adapter
	adapter, err := GetAdapter(config.Platform)
	if err != nil {
		return &SendResult{
			Status:   "failed",
			RespBody: err.Error(),
			Error:    err,
		}, err
	}

	// Format message
	body, err := adapter.FormatMessage(msg, raw, config)
	if err != nil {
		return &SendResult{
			Status:   "failed",
			RespBody: fmt.Sprintf("format message error: %v", err),
			Error:    err,
		}, err
	}

	// Prepare URL with signature if needed
	finalURL := config.TargetURL
	if config.HasSignature() {
		signedURL, sErr := adapter.SignRequest(config.TargetURL, body, config.SignSecret)
		if sErr != nil {
			return &SendResult{
				Status:   "failed",
				RespBody: fmt.Sprintf("sign request error: %v", sErr),
				Error:    sErr,
			}, sErr
		}
		finalURL = signedURL
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", finalURL, bytes.NewReader(body))
	if err != nil {
		return &SendResult{
			Status:   "failed",
			RespBody: fmt.Sprintf("create request error: %v", err),
			Error:    err,
		}, err
	}

	// Set headers
	req.Header.Set("Content-Type", adapter.GetContentType())
	req.Header.Set("User-Agent", "k8m-webhook-client/1.0")

	// Create a logged client with specific receiver info for this request
	loggedClient := NewLoggedHTTPClient(c.timeout, config.WebhookId, config.WebhookName, config.Platform)

	// Send request
	resp, webhookLog, err := loggedClient.DoWithLogging(req)
	if err != nil {
		return &SendResult{
			Status:   "failed",
			RespBody: fmt.Sprintf("send request error: %v", err),
			Error:    err,
		}, err
	}
	defer resp.Body.Close()

	// Log webhook details if available
	_ = webhookLog // webhookLog is already handled by LoggedHTTPClient

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		klog.Warningf("Failed to read response body: %v", err)
		respBody = []byte("failed to read response")
	}

	// Determine status
	status := "success"
	if resp.StatusCode >= 400 {
		status = "failed"
		err = fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	return &SendResult{
		Status:     status,
		StatusCode: resp.StatusCode,
		RespBody:   string(respBody),
		Error:      err,
	}, err
}
