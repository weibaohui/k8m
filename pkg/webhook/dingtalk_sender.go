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

// DingtalkSender implements webhook sending for Dingtalk.
type DingtalkSender struct{}

func (d *DingtalkSender) Name() string {
	return "dingtalk"
}

func (d *DingtalkSender) Send(msg string, receiver *Receiver) (*SendResult, error) {
	// Add Dingtalk signature if enabled
	finalURL := receiver.TargetURL
	if receiver.SignAlgo == "dingtalk" && receiver.SignSecret != "" {
		timestamp := time.Now().UnixMilli() // 钉钉使用毫秒时间戳
		timestampStr := strconv.FormatInt(timestamp, 10)
		signature, err := GenDingtalkSign(receiver.SignSecret, timestamp)
		if err != nil {
			return nil, err
		}

		params := url.Values{}
		params.Set("timestamp", timestampStr)
		params.Set("sign", signature)
		finalURL = fmt.Sprintf("%s?%s", finalURL, params.Encode())
	}

	payload := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": msg,
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

// GenDingtalkSign generates signature for Dingtalk webhook
func GenDingtalkSign(secret string, timestamp int64) (string, error) {
	// timestamp+"\n"+secret 做sha256, 再进行base64 encode
	stringToSign := fmt.Sprintf("%d", timestamp) + "\n" + secret
	h := hmac.New(sha256.New, []byte(secret))
	_, err := h.Write([]byte(stringToSign))
	if err != nil {
		return "", err
	}
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// 对签名进行URL编码
	urlEncodedSign := url.QueryEscape(signature)
	return urlEncodedSign, nil
}
