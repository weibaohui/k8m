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

func (f *FeishuSender) Send(msg string, receiver *Receiver) (*SendResult, error) {

	// Add Feishu signature if enabled
	finalURL := receiver.TargetURL
	if receiver.SignAlgo == "feishu" && receiver.SignSecret != "" {
		timestamp := time.Now().Unix()
		timestampStr := strconv.FormatInt(timestamp, 10)
		signature, err := GenSign(receiver.SignSecret, timestamp)
		if err != nil {
			return nil, err
		}
		params := url.Values{}
		params.Set("timestamp", timestampStr)
		params.Set("sign", signature)
		finalURL = fmt.Sprintf("%s?%s", finalURL, params.Encode())
	}
	payload := map[string]interface{}{
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
	for k, v := range receiver.Headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return &SendResult{Status: "failed"}, err
	}
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)

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
