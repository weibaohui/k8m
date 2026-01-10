package api

import "sync/atomic"

// SendResult 抽象 webhook 发送结果，避免调用方依赖具体 webhook 插件实现。
type SendResult struct {
	Status     string `json:"status"`
	StatusCode int    `json:"status_code"`
	RespBody   string `json:"resp_body"`
	Error      error  `json:"-"`
}

// Webhook 抽象 webhook 能力，对调用方隐藏具体插件实现和内部逻辑。
type Webhook interface {
	// PushMsgToAllTargetByIDs 中文函数注释：向指定接收者ID列表批量推送消息。
	PushMsgToAllTargetByIDs(msg string, raw string, receiverIDs []string) []*SendResult
	// GetNamesByIds 中文函数注释：根据接收者ID列表查询名称列表。
	GetNamesByIds(ids []string) ([]string, error)
}

// noopWebhook 为默认的空实现，保证在未注册真实实现时也不会产生空指针。
type noopWebhook struct{}

func (noopWebhook) PushMsgToAllTargetByIDs(msg string, raw string, receiverIDs []string) []*SendResult {
	return nil
}

func (noopWebhook) GetNamesByIds(ids []string) ([]string, error) {
	return []string{}, nil
}

var webhookVal atomic.Value // 保存 Webhook 实现，始终为非 nil

type webhookHolder struct {
	svc Webhook
}

func init() {
	webhookVal.Store(&webhookHolder{svc: noopWebhook{}})
}

// WebhookService 中文函数注释：返回当前生效的 Webhook 实现，始终非 nil。
func WebhookService() Webhook {
	return webhookVal.Load().(*webhookHolder).svc
}

// PushMsgToAllTargetByIDs 中文函数注释：向指定接收者ID列表批量推送消息（统一访问层便捷方法）。
func PushMsgToAllTargetByIDs(msg string, raw string, receiverIDs []string) []*SendResult {
	return WebhookService().PushMsgToAllTargetByIDs(msg, raw, receiverIDs)
}

// GetNamesByIds 中文函数注释：根据接收者ID列表查询名称列表（统一访问层便捷方法）。
func GetNamesByIds(ids []string) ([]string, error) {
	return WebhookService().GetNamesByIds(ids)
}

// RegisterWebhook 中文函数注释：在运行期注册或切换 Webhook 能力实现。
func RegisterWebhook(svc Webhook) {
	if svc == nil {
		svc = noopWebhook{}
	}
	webhookVal.Store(&webhookHolder{svc: svc})
}

// UnregisterWebhook 中文函数注释：在运行期取消注册 Webhook 能力，实现回退为 noop。
func UnregisterWebhook() {
	webhookVal.Store(&webhookHolder{svc: noopWebhook{}})
}
