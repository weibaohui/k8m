package webhook

import (
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/plugins/modules/webhook/core"
)

type SendResult = core.SendResult
type WebhookConfig = core.WebhookConfig
type PlatformAdapter = core.PlatformAdapter

var (
	ErrInvalidPlatform = core.ErrInvalidPlatform
	ErrInvalidURL      = core.ErrInvalidURL
	ErrSenderNotFound  = core.ErrSenderNotFound
	ErrSendFailed      = core.ErrSendFailed
	ErrInvalidConfig   = core.ErrInvalidConfig
)

func RegisterAllAdapters() { core.RegisterAllAdapters() }
func RegisterAdapter(platform string, adapter PlatformAdapter) {
	core.RegisterAdapter(platform, adapter)
}
func GetAdapter(platform string) (PlatformAdapter, error) { return core.GetAdapter(platform) }
func GetRegisteredPlatforms() []string                    { return core.GetRegisteredPlatforms() }

func NewWebhookConfig(receiver *models.WebhookReceiver) *WebhookConfig {
	return core.NewWebhookConfig(receiver)
}

func PushMsgToSingleTarget(msg string, raw string, receiver *models.WebhookReceiver) *SendResult {
	return core.PushMsgToSingleTarget(msg, raw, receiver)
}

func PushMsgToAllTargets(msg string, raw string, receivers []*models.WebhookReceiver) []*SendResult {
	return core.PushMsgToAllTargets(msg, raw, receivers)
}
