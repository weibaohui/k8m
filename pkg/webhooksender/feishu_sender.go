package webhooksender

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"html/template"
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

func (f *FeishuSender) Send(event *InspectionCheckEvent, receiver *WebhookReceiver) (*SendResult, error) {
	// Render content
	tmpl, err := template.New("payload").Parse(receiver.Template)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, event); err != nil {
		return nil, err
	}

	body := buf.Bytes()

	// Add Feishu signature if enabled
	finalURL := receiver.TargetURL
	if receiver.SignAlgo == "feishu" && receiver.SignSecret != "" {
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		stringToSign := timestamp + "\n" + receiver.SignSecret
		h := hmac.New(sha256.New, []byte(receiver.SignSecret))
		h.Write([]byte(stringToSign))
		signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
		params := url.Values{}
		params.Set("timestamp", timestamp)
		params.Set("sign", signature)
		finalURL = fmt.Sprintf("%s?%s", finalURL, params.Encode())
	}

	req, err := http.NewRequest("POST", finalURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range receiver.Headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 5 * time.Second}
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
