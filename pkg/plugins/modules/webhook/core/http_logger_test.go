package core

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLoggedHTTPClient(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"errcode": 0,
			"errmsg":  "ok",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer testServer.Close()

	client := NewLoggedHTTPClient(5*time.Second, 0, "test-sender", "test-receiver")

	requestBody := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": "测试消息",
		},
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req, err := http.NewRequest("POST", testServer.URL, bytes.NewReader(bodyBytes))
	if err != nil {
		t.Fatalf("创建请求失败: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, webhookLog, err := client.DoWithLogging(req)
	if err != nil {
		t.Fatalf("请求执行失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("期望状态码 200, 实际得到 %d", resp.StatusCode)
	}

	if webhookLog == nil {
		t.Fatal("webhook日志不应该为空")
	}

	if webhookLog.Request.Method != "POST" {
		t.Errorf("期望请求方法为 POST, 实际得到 %s", webhookLog.Request.Method)
	}

	if webhookLog.Request.WebhookName != "test-sender" {
		t.Errorf("期望发送器名称为 test-sender, 实际得到 %s", webhookLog.Request.WebhookName)
	}

	if webhookLog.Request.BodySize == 0 {
		t.Error("请求体大小不应该为0")
	}

	if webhookLog.Response.StatusCode != http.StatusOK {
		t.Errorf("期望响应状态码为 200, 实际得到 %d", webhookLog.Response.StatusCode)
	}

	if !webhookLog.Response.Success {
		t.Error("响应应该标记为成功")
	}

	if webhookLog.Response.Duration == 0 {
		t.Error("响应时间不应该为0")
	}

	if webhookLog.Summary == "" {
		t.Error("日志摘要不应该为空")
	}
}

func TestSanitizeURL(t *testing.T) {
	client := &LoggedHTTPClient{}

	testCases := []struct {
		input    string
		expected string
	}{
		{
			input:    "https://example.com/webhook?sign=abc123",
			expected: "https://example.com/webhook?sign=***",
		},
		{
			input:    "https://example.com/webhook?timestamp=123&signature=def456",
			expected: "https://example.com/webhook?timestamp=123&signature=***",
		},
		{
			input:    "https://example.com/webhook",
			expected: "https://example.com/webhook",
		},
	}

	for _, tc := range testCases {
		result := client.sanitizeURL(tc.input)
		if result != tc.expected {
			t.Errorf("sanitizeURL(%s) = %s, 期望 %s", tc.input, result, tc.expected)
		}
	}
}

func TestSanitizeHeader(t *testing.T) {
	client := &LoggedHTTPClient{}

	testCases := []struct {
		key      string
		value    string
		expected string
	}{
		{"Authorization", "Bearer token123", "***"},
		{"X-API-Key", "key123", "***"},
		{"Content-Type", "application/json", "application/json"},
		{"User-Agent", "test-agent", "test-agent"},
	}

	for _, tc := range testCases {
		result := client.sanitizeHeader(tc.key, tc.value)
		if result != tc.expected {
			t.Errorf("sanitizeHeader(%s, %s) = %s, 期望 %s", tc.key, tc.value, result, tc.expected)
		}
	}
}
