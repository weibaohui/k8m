package webhook

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"k8s.io/klog/v2"
)

// WebhookClient provides a unified HTTP transport layer for webhook sending.
type WebhookClient struct {
	httpClient *http.Client
	timeout    time.Duration
}

// NewWebhookClient creates a new webhook client with default settings.
func NewWebhookClient() *WebhookClient {
	return &WebhookClient{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		timeout: 30 * time.Second,
	}
}

// NewWebhookClientWithTimeout creates a new webhook client with custom timeout.
func NewWebhookClientWithTimeout(timeout time.Duration) *WebhookClient {
	return &WebhookClient{
		httpClient: &http.Client{
			Timeout: timeout,
		},
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
		signedURL, err := adapter.SignRequest(config.TargetURL, body, config.SignSecret)
		if err != nil {
			return &SendResult{
				Status:   "failed",
				RespBody: fmt.Sprintf("sign request error: %v", err),
				Error:    err,
			}, err
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

	// Send request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &SendResult{
			Status:   "failed",
			RespBody: fmt.Sprintf("send request error: %v", err),
			Error:    err,
		}, err
	}
	defer resp.Body.Close()

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

// SendToSingleTarget sends a message to a single webhook target.
// This is a convenience method that creates a WebhookConfig from WebhookReceiver.
func (c *WebhookClient) SendToSingleTarget(ctx context.Context, msg, raw string, receiver interface{}) (*SendResult, error) {
	var config *WebhookConfig
	
	// Handle different receiver types for backward compatibility
	switch r := receiver.(type) {
	case *WebhookConfig:
		config = r
	default:
		// Try to convert from models.WebhookReceiver if available
		// This would need to be implemented based on your models
		return nil, fmt.Errorf("unsupported receiver type: %T", receiver)
	}

	return c.Send(ctx, msg, raw, config)
}

// SendToMultipleTargets sends a message to multiple webhook targets concurrently.
func (c *WebhookClient) SendToMultipleTargets(ctx context.Context, msg, raw string, configs []*WebhookConfig) []*SendResult {
	results := make([]*SendResult, len(configs))
	
	// Use a channel to collect results
	type indexedResult struct {
		index  int
		result *SendResult
	}
	
	resultChan := make(chan indexedResult, len(configs))
	
	// Send to all targets concurrently
	for i, config := range configs {
		go func(index int, cfg *WebhookConfig) {
			result, _ := c.Send(ctx, msg, raw, cfg)
			resultChan <- indexedResult{index: index, result: result}
		}(i, config)
	}
	
	// Collect results
	for i := 0; i < len(configs); i++ {
		indexedRes := <-resultChan
		results[indexedRes.index] = indexedRes.result
		
		// Log result
		klog.V(6).Infof("Webhook sent to [%s] %s, result: [%s] status_code: %d", 
			configs[indexedRes.index].Platform, 
			configs[indexedRes.index].TargetURL, 
			indexedRes.result.Status,
			indexedRes.result.StatusCode)
	}
	
	return results
}

// buildSignedURL builds a URL with signature parameters.
func buildSignedURL(baseURL, timestamp, signature string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}
	
	params := u.Query()
	params.Set("timestamp", timestamp)
	params.Set("sign", signature)
	u.RawQuery = params.Encode()
	
	return u.String(), nil
}