// Package worker 实现事件处理Worker
package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/eventhandler/config"
	"github.com/weibaohui/k8m/pkg/eventhandler/watcher"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/webhook"
	"gorm.io/gorm"
	"k8s.io/klog/v2"
)

// EventWorker 事件处理Worker
type EventWorker struct {
	cfg          *config.EventHandlerConfig
	ruleMatcher  *watcher.RuleMatcher
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	processMutex sync.Mutex
}

// NewEventWorker 创建事件处理Worker
func NewEventWorker() *EventWorker {
	ctx, cancel := context.WithCancel(context.Background())
	cfg := config.DefaultEventHandlerConfig()

	return &EventWorker{
		cfg:         cfg,
		ruleMatcher: watcher.NewRuleMatcher(cfg.ClusterRules),
		ctx:         ctx,
		cancel:      cancel,
	}
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

// processBatch 处理一批事件
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

	for _, event := range k8sEvents {
		if err := w.processEvent(event); err != nil {
			klog.Errorf("处理事件失败: %v", err)
			// 增加重试次数
			if err := modelEvent.IncrementAttemptsByID(event.ID); err != nil {
				klog.Errorf("增加重试次数失败: %v", err)
			}
		}
	}

	return nil
}

// processEvent 处理单个事件
func (w *EventWorker) processEvent(event *models.K8sEvent) error {
	klog.V(6).Infof("处理事件: %s", event.EvtKey)

	// 检查重试次数
	if event.Attempts >= w.cfg.Worker.MaxRetries {
		klog.Warningf("事件达到最大重试次数，标记为已处理: %s", event.EvtKey)
		var m models.K8sEvent
		return m.MarkProcessedByID(event.ID, true)
	}

	// 按集群规则进行过滤；不匹配的直接标记为已处理，避免重复进入队列
	if !w.ruleMatcher.Match(event) {
		klog.V(6).Infof("事件不匹配集群规则，跳过推送: 集群=%s 键=%s", event.Cluster, event.EvtKey)
		var m models.K8sEvent
		return m.MarkProcessedByID(event.ID, true)
	}

	// 推送Webhook
	if err := w.pushWebhook(event); err != nil {
		klog.Errorf("Webhook推送失败: %v", err)
		// 推送失败不标记为已处理，让重试机制处理
		return err
	}

	// 标记为已处理
	var m models.K8sEvent
	return m.MarkProcessedByID(event.ID, true)
}

// pushWebhook 推送Webhook通知
func (w *EventWorker) pushWebhook(event *models.K8sEvent) error {
	// 按集群获取WebhookID列表
	webhookIDs, ok := w.cfg.Webhooks[event.Cluster]
	if !ok || len(webhookIDs) == 0 {
		klog.V(6).Infof("集群 %s 未配置Webhook，跳过推送", event.Cluster)
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
		klog.V(6).Infof("未配置webhook接收器，跳过推送")
		return nil
	}

	// 准备消息内容
	summary := fmt.Sprintf("Event Warning 事件\n集群：[%s]\n资源：%s/%s\n类型：%s\n原因：%s\n消息：%s",
		event.Cluster, event.Namespace, event.Name, event.Type, event.Reason, event.Message)
	resultRaw := utils.ToJSON(event)
	// 使用统一模式推送到所有目标
	results := webhook.PushMsgToAllTargets(summary, resultRaw, receivers)

	if len(results) > 0 && results[0].Error != nil {
		return fmt.Errorf("webhook推送失败: %w", results[0].Error)
	}

	klog.V(6).Infof("Webhook推送成功: %s", event.EvtKey)
	return nil
}
