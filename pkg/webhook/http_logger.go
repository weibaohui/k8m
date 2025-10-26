package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"k8s.io/klog/v2"
)

// HTTPRequestLog 记录HTTP请求的详细信息
type HTTPRequestLog struct {
	Timestamp   time.Time         `json:"timestamp"`
	Method      string            `json:"method"`
	URL         string            `json:"url"`
	Headers     map[string]string `json:"headers"`
	Body        string            `json:"body"`
	BodySize    int               `json:"body_size"`
	WebhookName string            `json:"webhook_name"`
	ReceiverID  string            `json:"receiver_id,omitempty"`
}

// HTTPResponseLog 记录HTTP响应的详细信息
type HTTPResponseLog struct {
	Timestamp    time.Time         `json:"timestamp"`
	StatusCode   int               `json:"status_code"`
	Status       string            `json:"status"`
	Headers      map[string]string `json:"headers"`
	Body         string            `json:"body"`
	BodySize     int               `json:"body_size"`
	Duration     time.Duration     `json:"duration"`
	Success      bool              `json:"success"`
	ErrorMessage string            `json:"error_message,omitempty"`
}

// WebhookLog 完整的webhook发送日志
type WebhookLog struct {
	Request  HTTPRequestLog  `json:"request"`
	Response HTTPResponseLog `json:"response"`
	Summary  string          `json:"summary"`
}

// LoggedHTTPClient 带日志记录功能的HTTP客户端包装器
type LoggedHTTPClient struct {
	client      *http.Client
	webhookId   uint
	webhookName string
	receiverID  string
}

// NewLoggedHTTPClient 创建一个新的带日志记录的HTTP客户端
func NewLoggedHTTPClient(timeout time.Duration, webhookId uint, webhookName, receiverID string) *LoggedHTTPClient {
	return &LoggedHTTPClient{
		client:      &http.Client{Timeout: timeout},
		webhookId:   webhookId,
		webhookName: webhookName,
		receiverID:  receiverID,
	}
}

// DoWithLogging 执行HTTP请求并记录详细日志
func (c *LoggedHTTPClient) DoWithLogging(req *http.Request) (*http.Response, *WebhookLog, error) {
	startTime := time.Now()

	// 记录请求信息
	requestLog := c.logRequest(req, startTime)

	// 执行请求
	resp, err := c.client.Do(req)
	endTime := time.Now()
	duration := endTime.Sub(startTime)

	// 记录响应信息
	responseLog := c.logResponse(resp, err, endTime, duration)

	// 创建完整的webhook日志
	webhookLog := &WebhookLog{
		Request:  requestLog,
		Response: responseLog,
		Summary:  c.generateSummary(requestLog, responseLog),
	}

	// 输出日志
	c.outputLog(webhookLog)

	return resp, webhookLog, err
}

// logRequest 记录HTTP请求详情
func (c *LoggedHTTPClient) logRequest(req *http.Request, timestamp time.Time) HTTPRequestLog {
	// 读取请求体（需要重新设置以供后续使用）
	var bodyContent string
	var bodySize int
	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err == nil {
			bodyContent = string(bodyBytes)
			bodySize = len(bodyBytes)
			// 重新设置请求体
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}
	}

	// 收集请求头（脱敏处理）
	headers := make(map[string]string)
	for key, values := range req.Header {
		if len(values) > 0 {
			headers[key] = c.sanitizeHeader(key, values[0])
		}
	}

	// 脱敏URL中的敏感信息
	sanitizedURL := c.sanitizeURL(req.URL.String())

	return HTTPRequestLog{
		Timestamp:   timestamp,
		Method:      req.Method,
		URL:         sanitizedURL,
		Headers:     headers,
		Body:        c.sanitizeBody(bodyContent),
		BodySize:    bodySize,
		WebhookName: c.webhookName,
		ReceiverID:  c.receiverID,
	}
}

// logResponse 记录HTTP响应详情
func (c *LoggedHTTPClient) logResponse(resp *http.Response, err error, timestamp time.Time, duration time.Duration) HTTPResponseLog {
	responseLog := HTTPResponseLog{
		Timestamp: timestamp,
		Duration:  duration,
		Success:   err == nil && resp != nil && resp.StatusCode < 400,
	}

	if err != nil {
		responseLog.ErrorMessage = err.Error()
		responseLog.StatusCode = 0
		responseLog.Status = "request_failed"
		return responseLog
	}

	if resp == nil {
		responseLog.ErrorMessage = "response is nil"
		responseLog.StatusCode = 0
		responseLog.Status = "no_response"
		return responseLog
	}

	// 记录响应基本信息
	responseLog.StatusCode = resp.StatusCode
	responseLog.Status = resp.Status

	// 收集响应头
	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	responseLog.Headers = headers

	// 读取响应体（需要重新设置以供后续使用）
	if resp.Body != nil {
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr == nil {
			responseLog.Body = string(bodyBytes)
			responseLog.BodySize = len(bodyBytes)
			// 重新设置响应体
			resp.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		} else {
			responseLog.ErrorMessage = fmt.Sprintf("failed to read response body: %v", readErr)
		}
	}

	return responseLog
}

// sanitizeURL 脱敏URL中的敏感信息
func (c *LoggedHTTPClient) sanitizeURL(url string) string {
	// 使用正则表达式脱敏签名参数
	// 匹配 sign=xxx 或 signature=xxx 格式，将值替换为 ***
	if strings.Contains(url, "sign=") {
		// 查找 sign= 后面的值并替换
		parts := strings.Split(url, "sign=")
		if len(parts) > 1 {
			// 找到第一个 & 或字符串结尾
			valueAndRest := parts[1]
			ampIndex := strings.Index(valueAndRest, "&")
			if ampIndex != -1 {
				url = parts[0] + "sign=***&" + valueAndRest[ampIndex+1:]
			} else {
				url = parts[0] + "sign=***"
			}
		}
	}
	if strings.Contains(url, "signature=") {
		// 查找 signature= 后面的值并替换
		parts := strings.Split(url, "signature=")
		if len(parts) > 1 {
			// 找到第一个 & 或字符串结尾
			valueAndRest := parts[1]
			ampIndex := strings.Index(valueAndRest, "&")
			if ampIndex != -1 {
				url = parts[0] + "signature=***&" + valueAndRest[ampIndex+1:]
			} else {
				url = parts[0] + "signature=***"
			}
		}
	}
	return url
}

// sanitizeHeader 脱敏请求头中的敏感信息
func (c *LoggedHTTPClient) sanitizeHeader(key, value string) string {
	lowerKey := strings.ToLower(key)
	if strings.Contains(lowerKey, "authorization") ||
		strings.Contains(lowerKey, "token") ||
		strings.Contains(lowerKey, "secret") ||
		strings.Contains(lowerKey, "key") {
		return "***"
	}
	return value
}

// sanitizeBody 脱敏请求体中的敏感信息
func (c *LoggedHTTPClient) sanitizeBody(body string) string {
	// 如果是JSON格式，尝试解析并脱敏
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(body), &jsonData); err == nil {
		// 脱敏可能的敏感字段
		for key, value := range jsonData {
			lowerKey := strings.ToLower(key)
			if strings.Contains(lowerKey, "secret") ||
				strings.Contains(lowerKey, "token") ||
				strings.Contains(lowerKey, "password") ||
				strings.Contains(lowerKey, "key") {
				jsonData[key] = "***"
			} else if strValue, ok := value.(string); ok && len(strValue) > 100 {
				// 截断过长的字符串
				jsonData[key] = strValue[:100] + "..."
			}
		}
		if sanitizedBytes, err := json.Marshal(jsonData); err == nil {
			return string(sanitizedBytes)
		}
	}

	// 如果不是JSON或解析失败，直接截断过长内容
	if len(body) > 1000 {
		return body[:1000] + "..."
	}
	return body
}

// generateSummary 生成日志摘要
func (c *LoggedHTTPClient) generateSummary(req HTTPRequestLog, resp HTTPResponseLog) string {
	status := "SUCCESS"
	if !resp.Success {
		status = "FAILED"
	}

	return fmt.Sprintf("[%d-%s] %s %s -> %d %s (%.2fms)",
		c.webhookId,
		c.webhookName,
		req.Method,
		req.URL,
		resp.StatusCode,
		status,
		float64(resp.Duration.Nanoseconds())/1e6,
	)
}

// outputLog 输出日志到不同级别
func (c *LoggedHTTPClient) outputLog(log *WebhookLog) {
	// 输出摘要到INFO级别
	klog.Infof("Webhook Send: %s", log.Summary)

	// 输出详细信息到V(6)级别
	if klog.V(6).Enabled() {
		if logBytes, err := json.MarshalIndent(log, "", "  "); err == nil {
			klog.V(6).Infof("Webhook Detail Log:\n%s", string(logBytes))
		}
	}

	// 如果发送失败，输出错误信息
	if !log.Response.Success {
		klog.Errorf("Webhook Send Failed: %s, Error: %s, Response: %s",
			log.Summary,
			log.Response.ErrorMessage,
			log.Response.Body,
		)
	}
}
