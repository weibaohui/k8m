package webhook

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/weibaohui/htpl"
	"k8s.io/klog/v2"
)

// DingtalkAdapter implements PlatformAdapter for DingTalk webhooks.
type DingtalkAdapter struct{}

func (d *DingtalkAdapter) Name() string {
	return "dingtalk"
}

func (d *DingtalkAdapter) GetContentType() string {
	return "application/json"
}

func (d *DingtalkAdapter) FormatMessage(msg, raw string, config *WebhookConfig) ([]byte, error) {
	payload := map[string]any{
		"msgtype": "text",
		"text": map[string]string{
			"content": msg,
		},
	}
	return json.Marshal(payload)
}

func (d *DingtalkAdapter) SignRequest(baseURL string, body []byte, secret string) (string, error) {
	if secret == "" {
		return baseURL, nil
	}

	timestamp := time.Now().UnixNano() / 1e6 // DingTalk uses millisecond timestamp
	timestampStr := strconv.FormatInt(timestamp, 10)
	
	signature, err := d.generateSignature(secret, timestamp)
	if err != nil {
		return "", err
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	params := u.Query()
	params.Set("timestamp", timestampStr)
	params.Set("sign", signature)
	u.RawQuery = params.Encode()

	return u.String(), nil
}

func (d *DingtalkAdapter) generateSignature(secret string, timestamp int64) (string, error) {
	timestampStr := strconv.FormatInt(timestamp, 10)
	stringToSign := timestampStr + "\n" + secret
	
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	
	return url.QueryEscape(signature), nil
}

// FeishuAdapter implements PlatformAdapter for Feishu webhooks.
type FeishuAdapter struct{}

func (f *FeishuAdapter) Name() string {
	return "feishu"
}

func (f *FeishuAdapter) GetContentType() string {
	return "application/json"
}

func (f *FeishuAdapter) FormatMessage(msg, raw string, config *WebhookConfig) ([]byte, error) {
	payload := map[string]any{
		"msg_type": "text",
		"content": map[string]string{
			"text": msg,
		},
	}
	return json.Marshal(payload)
}

func (f *FeishuAdapter) SignRequest(baseURL string, body []byte, secret string) (string, error) {
	if secret == "" {
		return baseURL, nil
	}

	timestamp := time.Now().Unix()
	timestampStr := strconv.FormatInt(timestamp, 10)
	
	signature, err := f.generateSignature(secret, timestamp)
	if err != nil {
		return "", err
	}

	u, err := url.Parse(baseURL)
	if err != nil {
		return "", err
	}

	params := u.Query()
	params.Set("timestamp", timestampStr)
	params.Set("sign", signature)
	u.RawQuery = params.Encode()

	return u.String(), nil
}

func (f *FeishuAdapter) generateSignature(secret string, timestamp int64) (string, error) {
	timestampStr := strconv.FormatInt(timestamp, 10)
	stringToSign := timestampStr + "\n" + secret
	
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	
	return signature, nil
}

// WechatAdapter implements PlatformAdapter for WeChat Work webhooks.
type WechatAdapter struct{}

func (w *WechatAdapter) Name() string {
	return "wechat"
}

func (w *WechatAdapter) GetContentType() string {
	return "application/json"
}

func (w *WechatAdapter) FormatMessage(msg, raw string, config *WebhookConfig) ([]byte, error) {
	payload := map[string]any{
		"msgtype": "markdown",
		"markdown": map[string]string{
			"content": msg,
		},
	}
	return json.Marshal(payload)
}

func (w *WechatAdapter) SignRequest(baseURL string, body []byte, secret string) (string, error) {
	// WeChat Work doesn't require signing, return URL as-is
	return baseURL, nil
}

// DefaultAdapter implements PlatformAdapter for custom/generic webhooks.
type DefaultAdapter struct{}

func (d *DefaultAdapter) Name() string {
	return "default"
}

func (d *DefaultAdapter) GetContentType() string {
	return "application/json"
}

func (d *DefaultAdapter) FormatMessage(msg, raw string, config *WebhookConfig) ([]byte, error) {
	template := config.GetEffectiveTemplate()
	
	// If no template, return plain message
	if template == "" || template == "%s" {
		return []byte(msg), nil
	}

	// Use htpl for template rendering
	bodyTemplate := strings.ReplaceAll(template, "{{msg}}", "${msg}")
	bodyTemplate = strings.ReplaceAll(bodyTemplate, "{{raw}}", "${raw}")
	
	eng := htpl.NewEngine()
	tpl, err := eng.ParseString(bodyTemplate)
	if err != nil {
		klog.Errorf("Webhook DefaultAdapter template parse failed: platform=%s target=%s template=%q error=%v", 
			config.Platform, config.TargetURL, bodyTemplate, err)
		return []byte(msg), nil // Fallback to plain message
	}

	ctx := map[string]any{
		"msg": msg,
		"raw": raw,
	}
	
	rendered, err := tpl.Render(ctx)
	if err != nil {
		klog.Errorf("Webhook DefaultAdapter template render failed: platform=%s target=%s template=%q error=%v", 
			config.Platform, config.TargetURL, bodyTemplate, err)
		return []byte(msg), nil // Fallback to plain message
	}

	if rendered == "" {
		return []byte(msg), nil
	}

	return []byte(rendered), nil
}

func (d *DefaultAdapter) SignRequest(baseURL string, body []byte, secret string) (string, error) {
	// Default adapter doesn't implement signing
	// Custom signing should be implemented by specific adapters
	return baseURL, nil
}