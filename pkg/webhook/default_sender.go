package webhook

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"time"
)

// DefaultSender sends JSON webhook with optional HMAC-SHA256 signature.
type DefaultSender struct{}

func (d *DefaultSender) Name() string {
	return "default"
}

func (d *DefaultSender) Send(msg string, receiver *Receiver) (*SendResult, error) {

	req, err := http.NewRequest(receiver.Method, receiver.TargetURL, bytes.NewReader([]byte(msg)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range receiver.Headers {
		req.Header.Set(k, v)
	}

	if receiver.SignAlgo == "hmac-sha256" && receiver.SignSecret != "" {
		h := hmac.New(sha256.New, []byte(receiver.SignSecret))
		h.Write([]byte(msg))
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
