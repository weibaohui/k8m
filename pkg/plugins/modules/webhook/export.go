package webhook

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/webhook/core"
	"github.com/weibaohui/k8m/pkg/plugins/modules/webhook/models"
	"k8s.io/klog/v2"
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
	// 检查插件是否已启用
	if !plugins.ManagerInstance().IsEnabled(modules.PluginNameWebhook) {
		klog.V(4).Infof("webhook 插件已禁用，跳过向单个接收者发送消息 %s", receiver.Name)
		return nil
	}
	return core.PushMsgToSingleTarget(msg, raw, receiver)
}

func PushMsgToAllTargets(msg string, raw string, receivers []*models.WebhookReceiver) []*SendResult {
	// 检查插件是否已启用
	if !plugins.ManagerInstance().IsEnabled(modules.PluginNameWebhook) {
		klog.V(4).Infof("webhook 插件已禁用，跳过向 %d 个接收者发送消息", len(receivers))
		return nil
	}

	return core.PushMsgToAllTargets(msg, raw, receivers)
}
