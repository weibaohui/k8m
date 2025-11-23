// Package worker 实现事件处理Worker
package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/eventhandler/config"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/webhook"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

// EventWorker 事件处理Worker
type EventWorker struct {
	cfg          *config.EventHandlerConfig
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	processMutex sync.Mutex
}

var defaultWorker *EventWorker

// NewEventWorker 创建事件处理Worker
func NewEventWorker() *EventWorker {
	ctx, cancel := context.WithCancel(context.Background())
	cfg := config.DefaultEventHandlerConfig()

	ew := &EventWorker{
		cfg:    cfg,
		ctx:    ctx,
		cancel: cancel,
	}
	// 注册为全局实例，便于控制器更新配置后即时生效
	defaultWorker = ew
	return ew
}

// Start 启动Worker
func (w *EventWorker) Start() {
	if w.cfg.Enabled {
		klog.V(6).Infof("启动事件处理Worker")
		w.wg.Add(1)
		go w.processLoop()
	} else {
		klog.V(6).Infof("事件转发功能未开启")
	}

}

// Stop 停止Worker
func (w *EventWorker) Stop() {
	if w.cfg.Enabled {
		klog.V(6).Infof("停止事件处理Worker")
		w.cancel()
		w.wg.Wait()
	}

}

// processLoop 处理循环
func (w *EventWorker) processLoop() {
	defer w.wg.Done()

	ticker := time.NewTicker(time.Duration(w.cfg.Worker.ProcessInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			if err := w.processBatch(); err != nil {
				klog.Errorf("处理事件批次失败: %v", err)
			}
		}
	}
}

// UpdateConfig 动态刷新事件处理配置（从数据库重新加载）
// 中文函数注释：用于在管理界面更新事件规则或Webhook后，立即让Worker生效，无需重启。
func (w *EventWorker) UpdateConfig() {
	if w == nil {
		return
	}
	// 重新加载配置
	newCfg := config.DefaultEventHandlerConfig()
	if newCfg == nil {
		return
	}
	// 原子更新配置与匹配器
	w.processMutex.Lock()
	w.cfg = newCfg
	w.processMutex.Unlock()
	klog.V(6).Infof("事件处理配置已更新，立即生效")
}

// Instance 获取全局事件处理Worker实例
// 中文函数注释：用于控制器在保存配置后调用刷新方法。
func Instance() *EventWorker {
	return defaultWorker
}

// processBatch 按批次获取未处理事件，逐条按每条事件配置进行过滤与推送
// 中文函数注释：针对数据库中的每一条 K8sEventConfig 规则，动态读取其集群与 Webhook 设置，
// 针对本批次未处理事件进行匹配与推送，直到遍历完所有规则；同一事件在成功推送后将被标记为已处理，避免重复推送。
func (w *EventWorker) processBatch() error {
	w.processMutex.Lock()
	defer w.processMutex.Unlock()

	// 获取未处理的事件（通过模型方法）
	var modelEvent models.K8sEvent
	k8sEvents, err := modelEvent.ListUnprocessed(w.cfg.Worker.BatchSize)
	if err != nil {
		return fmt.Errorf("获取未处理事件失败: %w", err)
	}

	if len(k8sEvents) == 0 {
		return nil
	}

	klog.V(6).Infof("开始处理事件批次: %d个事件", len(k8sEvents))

	// 本轮处理中已成功处理的事件ID，避免重复推送
	processedIDs := make(map[int64]bool)
	// 本轮中曾被任何规则匹配到的事件ID（用于最终未匹配的直接标记处理）
	matchedIDs := make(map[int64]bool)

	// 遍历每一条事件配置规则
	for _, ec := range w.cfg.EventConfigs {
		// 解析当前规则的集群列表与 webhook 列表
		clusters := make(map[string]struct{})
		for _, c := range strings.Split(ec.Clusters, ",") {
			cc := strings.TrimSpace(c)
			if cc != "" {
				clusters[cc] = struct{}{}
			}
		}
		var webhookIDs []string
		for _, wid := range strings.Split(ec.Webhooks, ",") {
			wtrim := strings.TrimSpace(wid)
			if wtrim != "" {
				webhookIDs = append(webhookIDs, wtrim)
			}
		}

		// 解析规则 JSON 字段为 RuleConfig
		var namespaces, names, labels, reasons, types []string
		if ec.RuleNamespaces != "" {
			_ = json.Unmarshal([]byte(ec.RuleNamespaces), &namespaces)
		}
		if ec.RuleNames != "" { // 目前不参与匹配，但保留解析
			_ = json.Unmarshal([]byte(ec.RuleNames), &names)
		}
		if ec.RuleLabels != "" { // 目前不参与匹配，但保留解析
			_ = json.Unmarshal([]byte(ec.RuleLabels), &labels)
		}
		if ec.RuleReasons != "" {
			_ = json.Unmarshal([]byte(ec.RuleReasons), &reasons)
		}
		if ec.RuleTypes != "" {
			_ = json.Unmarshal([]byte(ec.RuleTypes), &types)
		}
		rule := config.RuleConfig{Namespaces: namespaces, Labels: labels, Reasons: reasons, Types: types, Reverse: ec.RuleReverse}

		// 逐条事件，按当前规则筛选并按集群分组
		grouped := make(map[string][]*models.K8sEvent)
		for _, event := range k8sEvents {
			if processedIDs[event.ID] {
				continue
			}
			// 超过最大重试次数，直接标记为已处理
			if event.Attempts >= w.cfg.Worker.MaxRetries {
				klog.Warningf("事件达到最大重试次数，标记为已处理: %s", event.EvtKey)
				var m models.K8sEvent
				if err := m.MarkProcessedByID(event.ID, true); err != nil {
					klog.Errorf("标记事件已处理失败: %v", err)
				}
				processedIDs[event.ID] = true
				continue
			}
			// 集群限定：仅处理当前规则所配置的集群
			if _, ok := clusters[event.Cluster]; !ok {
				continue
			}
			// 使用规则匹配器进行匹配（按当前事件的集群为键）
			matcher := NewRuleMatcher(map[string]config.RuleConfig{event.Cluster: rule})
			if !matcher.Match(event) {
				continue
			}
			matchedIDs[event.ID] = true
			grouped[event.Cluster] = append(grouped[event.Cluster], event)
		}

		if len(grouped) == 0 {
			continue
		}

		// 按当前规则的 webhookIDs 进行批量推送
		for cluster, events := range grouped {
			if err := w.pushWebhookBatchForIDs(cluster, webhookIDs, events, ec.Name); err != nil {
				klog.Errorf("批量Webhook推送失败: 规则=%s 集群=%s 错误=%v", ec.Name, cluster, err)
				for _, e := range events {
					if err := modelEvent.IncrementAttemptsByID(e.ID); err != nil {
						klog.Errorf("增加重试次数失败: %v", err)
					}
				}
			} else {
				var m models.K8sEvent
				for _, e := range events {
					if err := m.MarkProcessedByID(e.ID, true); err != nil {
						klog.Errorf("标记事件已处理失败: %v", err)
					} else {
						processedIDs[e.ID] = true
					}
				}
			}
		}
	}

	// 对未被任何规则匹配的事件，直接标记为已处理，避免反复检查
	for _, event := range k8sEvents {
		if !processedIDs[event.ID] && !matchedIDs[event.ID] {
			var m models.K8sEvent
			if err := m.MarkProcessedByID(event.ID, true); err != nil {
				klog.Errorf("标记事件已处理失败: %v", err)
			}
		}
	}

	return nil

}

// pushWebhookBatchForIDs 根据指定的 webhookID 列表批量推送事件
// 中文函数注释：用于按单条事件配置指定的 webhook 目标推送消息，不再依赖全局按集群的 webhook 映射。
func (w *EventWorker) pushWebhookBatchForIDs(cluster string, webhookIDs []string, events []*models.K8sEvent, ruleName string) error {
	if len(webhookIDs) == 0 {
		klog.V(6).Infof("规则 %s 未配置Webhook，跳过推送", ruleName)
		return nil
	}

	// 查询所有已配置的Webhook接收器
	receiver := &models.WebhookReceiver{}
	receivers, _, err := receiver.List(dao.BuildDefaultParams(), func(d *gorm.DB) *gorm.DB {
		return d.Where("id IN ?", webhookIDs)
	})
	if err != nil {
		return fmt.Errorf("查询webhook接收器失败: %w", err)
	}
	if len(receivers) == 0 {
		klog.V(6).Infof("规则 %s 未找到可用的webhook接收器，跳过推送", ruleName)
		return nil
	}

	// 生成批量摘要与原始JSON数组
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Event Warning 批量事件\n规则：[%s]\n集群：[%s]\n数量：%d\n\n", ruleName, cluster, len(events)))
	for _, e := range events {
		sb.WriteString(fmt.Sprintf("资源：%s/%s\n类型：%s\n原因：%s\n消息：%s\n时间：%s\n\n",
			e.Namespace, e.Name, e.Type, e.Reason, e.Message, e.Timestamp.Format("2006-01-02 15:04:05")))
	}
	summary := sb.String()
	resultRaw := utils.ToJSON(events)

	// 使用统一模式推送到所有目标
	results := webhook.PushMsgToAllTargets(summary, resultRaw, receivers)

	// 判断是否全部失败：至少一个成功则认为成功
	allFailed := true
	for _, r := range results {
		if r != nil && r.Status == "success" && r.Error == nil {
			allFailed = false
			break
		}
	}
	if allFailed {
		return fmt.Errorf("批量webhook推送全部失败")
	}

	klog.V(6).Infof("批量Webhook推送成功: 规则=%s 集群=%s 事件数=%d", ruleName, cluster, len(events))
	return nil
}
