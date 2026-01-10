package api

import (
	"sync/atomic"

	"k8s.io/klog/v2"
)

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
	klog.V(4).Infof("Webhook 插件未开启,PushMsgToAllTargetByID 方法未执行 ")
	return nil
}

func (noopWebhook) GetNamesByIds(ids []string) ([]string, error) {
	klog.V(4).Infof("Webhook 插件未开启,GetNamesById 方法未执行")
	return []string{}, nil
}

var webhookVal atomic.Value // 保存 Webhook 实现，始终为非 nil

type webhookHolder struct {
	svc Webhook
}

func initWebhookNoop() {
	webhookVal.Store(&webhookHolder{svc: noopWebhook{}})
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
