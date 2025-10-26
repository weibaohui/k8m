package webhook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// FeishuSender implements webhook sending for Feishu.
type FeishuSender struct{}

func (f *FeishuSender) Name() string {
	return "feishu"
}

func (f *FeishuSender) Send(msg string, raw string, channel *Channel) (*SendResult, error) {

	// Add Feishu signature if enabled
	finalURL := channel.TargetURL
	if channel.SignSecret != "" {
		timestamp := time.Now().Unix()
		timestampStr := strconv.FormatInt(timestamp, 10)
		signature, err := GenSign(channel.SignSecret, timestamp)
		if err != nil {
			return nil, err
		}
		params := url.Values{}
		params.Set("timestamp", timestampStr)
		params.Set("sign", signature)
		finalURL = fmt.Sprintf("%s?%s", finalURL, params.Encode())
	}
	payload := map[string]any{
		"msg_type": "text",
		"content": map[string]string{
			"text": msg,
		},
	}
	msgBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", finalURL, bytes.NewReader(msgBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	// 使用带日志记录的HTTP客户端
	loggedClient := NewLoggedHTTPClient(60*time.Second, "feishu", channel.TargetURL)
	resp, webhookLog, err := loggedClient.DoWithLogging(req)
	if err != nil {
		return &SendResult{Status: "failed"}, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

	// 记录webhook日志到结果中（可选，用于后续查询）
	_ = webhookLog

	status := "success"
	if resp.StatusCode >= 400 {
		status = "failed"
	}
	return &SendResult{
		Status:     status,
		StatusCode: resp.StatusCode,
		RespBody:   string(respBody),
	}, nil
}
func GenSign(secret string, timestamp int64) (string, error) {
	// timestamp + key 做sha256, 再进行base64 encode
	stringToSign := fmt.Sprintf("%v", timestamp) + "\n" + secret
	var data []byte
	h := hmac.New(sha256.New, []byte(stringToSign))
	_, err := h.Write(data)
	if err != nil {
		return "", err
	}
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return signature, nil
}
