package webhooksender

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"html/template"
	"io"
	"net/http"
	"time"
)

// DefaultSender sends JSON webhook with optional HMAC-SHA256 signature.
type DefaultSender struct{}

func (d *DefaultSender) Name() string {
	return "default"
}

func (d *DefaultSender) Send(event *InspectionCheckEvent, receiver *WebhookReceiver) (*SendResult, error) {
	tmpl, err := template.New("payload").Parse(receiver.Template)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, event); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(receiver.Method, receiver.TargetURL, bytes.NewReader(buf.Bytes()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range receiver.Headers {
		req.Header.Set(k, v)
	}

	if receiver.SignAlgo == "hmac-sha256" && receiver.SignSecret != "" {
		h := hmac.New(sha256.New, []byte(receiver.SignSecret))
		h.Write(buf.Bytes())
		signature := hex.EncodeToString(h.Sum(nil))
		req.Header.Set(receiver.SignHeaderKey, signature)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return &SendResult{Status: "failed"}, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	status := "success"
	if resp.StatusCode >= 400 {
		status = "failed"
	}
	return &SendResult{
		Status:     status,
		StatusCode: resp.StatusCode,
		RespBody:   string(body),
	}, nil
}
