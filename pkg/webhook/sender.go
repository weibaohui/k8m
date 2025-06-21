package webhook

import (
	"fmt"
)

// SendResult holds the result of a webhook send attempt.
type SendResult struct {
	Status     string // success / failed
	StatusCode int
	RespBody   string
}

// Sender defines the webhook adapter interface.
type Sender interface {
	Name() string
	Send(msg string, receiver *Receiver) (*SendResult, error)
}

// senderRegistry holds all registered senders.
var senderRegistry = map[string]Sender{}

// RegisterSender registers a new platform sender.
func RegisterSender(platform string, sender Sender) {
	senderRegistry[platform] = sender
}

// getSender returns the appropriate sender.
// 若找不到平台，返回 nil 并建议调用方处理异常
func getSender(platform string) (Sender, error) {
	if s, ok := senderRegistry[platform]; ok {
		return s, nil
	}
	return nil, fmt.Errorf("webhook sender for platform '%s' not found", platform)
}
