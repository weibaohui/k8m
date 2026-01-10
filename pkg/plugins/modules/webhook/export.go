package webhook

import (
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/api"
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
	if !plugins.ManagerInstance().IsRunning(modules.PluginNameWebhook) {
		klog.V(4).Infof("webhook 插件已禁用，跳过向单个接收者发送消息 %s", receiver.Name)
		return nil
	}
	return core.PushMsgToSingleTarget(msg, raw, receiver)
}

func PushMsgToAllTargets(msg string, raw string, receivers []*models.WebhookReceiver) []*SendResult {
	// 检查插件是否已启用
	if !plugins.ManagerInstance().IsRunning(modules.PluginNameWebhook) {
		klog.V(4).Infof("webhook 插件已禁用，跳过向 %d 个接收者发送消息", len(receivers))
		return nil
	}

	return core.PushMsgToAllTargets(msg, raw, receivers)
}

func PushMsgToAllTargetByIDs(msg string, raw string, receiverIDs []string) []*SendResult {
	// 检查插件是否已启用
	if !plugins.ManagerInstance().IsRunning(modules.PluginNameWebhook) {
		klog.V(4).Infof("webhook 插件已禁用，跳过向 %d 个接收者发送消息", len(receiverIDs))
		return nil
	}

	return core.PushMsgToAllTargetByIDs(msg, raw, receiverIDs)
}

func GetNamesByIds(ids []string) ([]string, error) {
	// 检查插件是否已启用
	if !plugins.ManagerInstance().IsRunning(modules.PluginNameWebhook) {
		klog.V(4).Info("webhook 插件已禁用，返回空列表")
		return []string{}, nil
	}
	webhookReceiver := models.WebhookReceiver{}
	return webhookReceiver.GetNamesByIds(ids)
}

type webhookAPIService struct{}

// PushMsgToAllTargetByIDs 中文函数注释：向指定接收者ID列表批量推送消息（统一访问层实现）。
func (webhookAPIService) PushMsgToAllTargetByIDs(msg string, raw string, receiverIDs []string) []*api.SendResult {
	if !plugins.ManagerInstance().IsRunning(modules.PluginNameWebhook) {
		klog.V(4).Infof("webhook 插件已禁用，跳过向 %d 个接收者发送消息", len(receiverIDs))
		return nil
	}
	return toAPISendResults(core.PushMsgToAllTargetByIDs(msg, raw, receiverIDs))
}

// GetNamesByIds 中文函数注释：根据接收者ID列表查询名称列表（统一访问层实现）。
func (webhookAPIService) GetNamesByIds(ids []string) ([]string, error) {
	if !plugins.ManagerInstance().IsRunning(modules.PluginNameWebhook) {
		klog.V(4).Info("webhook 插件已禁用，返回空列表")
		return []string{}, nil
	}
	webhookReceiver := models.WebhookReceiver{}
	return webhookReceiver.GetNamesByIds(ids)
}

func toAPISendResult(r *core.SendResult) *api.SendResult {
	if r == nil {
		return nil
	}
	return &api.SendResult{
		Status:     r.Status,
		StatusCode: r.StatusCode,
		RespBody:   r.RespBody,
		Error:      r.Error,
	}
}

func toAPISendResults(results []*core.SendResult) []*api.SendResult {
	if len(results) == 0 {
		return nil
	}
	out := make([]*api.SendResult, 0, len(results))
	for _, r := range results {
		out = append(out, toAPISendResult(r))
	}
	return out
}
