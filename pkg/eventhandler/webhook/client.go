// Package webhook 实现Webhook推送功能
package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/weibaohui/k8m/pkg/eventhandler/model"
	"k8s.io/klog/v2"
)

// WebhookClient Webhook客户端接口
type WebhookClient interface {
	// Push 推送事件
	Push(event *model.Event) error
	// PushBatch 批量推送事件
	PushBatch(events []*model.Event) error
}

// HTTPWebhookClient HTTP Webhook客户端
type HTTPWebhookClient struct {
	config     *model.WebhookConfig
	httpClient *http.Client
}

// NewHTTPWebhookClient 创建HTTP Webhook客户端
func NewHTTPWebhookClient(config *model.WebhookConfig) *HTTPWebhookClient {
	if config.Timeout == 0 {
		config.Timeout = 30
	}
	if config.Retries == 0 {
		config.Retries = 3
	}

	return &HTTPWebhookClient{
		config: config,
		httpClient: &http.Client{
			Timeout: time.Duration(config.Timeout) * time.Second,
		},
	}
}

// Push 推送单个事件
func (c *HTTPWebhookClient) Push(event *model.Event) error {
	return c.PushBatch([]*model.Event{event})
}

// PushBatch 批量推送事件
func (c *HTTPWebhookClient) PushBatch(events []*model.Event) error {
	if len(events) == 0 {
		return nil
	}

	// 准备请求数据
	payload := map[string]interface{}{
		"events":    events,
		"count":     len(events),
		"timestamp": time.Now().Unix(),
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化事件数据失败: %w", err)
	}

	// 重试机制
	var lastErr error
	for i := 0; i <= c.config.Retries; i++ {
		if i > 0 {
			klog.V(6).Infof("Webhook推送重试第%d次", i)
			time.Sleep(time.Duration(i) * time.Second) // 指数退避
		}

		if err := c.doPush(data); err != nil {
			lastErr = err
			klog.Errorf("Webhook推送失败: %v", err)
			continue
		}

		klog.V(6).Infof("Webhook推送成功: %d个事件", len(events))
		return nil
	}

	return fmt.Errorf("Webhook推送失败(重试%d次): %w", c.config.Retries, lastErr)
}

// doPush 执行推送请求
func (c *HTTPWebhookClient) doPush(data []byte) error {
	req, err := http.NewRequest(c.config.Method, c.config.URL, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	for key, value := range c.config.Headers {
		req.Header.Set(key, value)
	}

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("发送请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("Webhook返回错误状态码: %d", resp.StatusCode)
	}

	return nil
}

// MockWebhookClient 模拟Webhook客户端（用于测试）
type MockWebhookClient struct {
	pushCount int
}

// NewMockWebhookClient 创建模拟Webhook客户端
func NewMockWebhookClient() *MockWebhookClient {
	return &MockWebhookClient{}
}

// Push 推送单个事件（模拟）
func (m *MockWebhookClient) Push(event *model.Event) error {
	m.pushCount++
	klog.V(6).Infof("模拟Webhook推送: %s", event.EvtKey)
	return nil
}

// PushBatch 批量推送事件（模拟）
func (m *MockWebhookClient) PushBatch(events []*model.Event) error {
	m.pushCount += len(events)
	klog.V(6).Infof("模拟Webhook批量推送: %d个事件", len(events))
	return nil
}

// GetPushCount 获取推送计数
func (m *MockWebhookClient) GetPushCount() int {
	return m.pushCount
}
