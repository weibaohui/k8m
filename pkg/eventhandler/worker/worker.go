// Package worker 实现事件处理Worker
package worker

import (
	"context"
	"fmt"
	"strings"
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

var defaultWorker *EventWorker

// NewEventWorker 创建事件处理Worker
func NewEventWorker() *EventWorker {
	ctx, cancel := context.WithCancel(context.Background())
	cfg := config.DefaultEventHandlerConfig()

	ew := &EventWorker{
		cfg:         cfg,
		ruleMatcher: watcher.NewRuleMatcher(cfg.ClusterRules),
		ctx:         ctx,
		cancel:      cancel,
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
	if w.ruleMatcher == nil {
		w.ruleMatcher = watcher.NewRuleMatcher(newCfg.ClusterRules)
	} else {
		w.ruleMatcher.UpdateRules(newCfg.ClusterRules)
	}
	w.processMutex.Unlock()
	klog.V(6).Infof("事件处理配置已更新，立即生效")
}

// Instance 获取全局事件处理Worker实例
// 中文函数注释：用于控制器在保存配置后调用刷新方法。
func Instance() *EventWorker {
	return defaultWorker
}

// processBatch 按批次获取未处理事件，逐条过滤并按集群分组后批量推送
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

	// 逐条过滤并按集群分组
	grouped := make(map[string][]*models.K8sEvent)
	for _, event := range k8sEvents {
		// 超过最大重试次数，直接标记为已处理
		if event.Attempts >= w.cfg.Worker.MaxRetries {
			klog.Warningf("事件达到最大重试次数，标记为已处理: %s", event.EvtKey)
			var m models.K8sEvent
			if err := m.MarkProcessedByID(event.ID, true); err != nil {
				klog.Errorf("标记事件已处理失败: %v", err)
			}
			continue
		}
		// 过滤不匹配的事件，直接标记为已处理
		if !w.ruleMatcher.Match(event) {
			klog.V(6).Infof("事件不匹配集群规则，跳过推送: 集群=%s 键=%s", event.Cluster, event.EvtKey)
			var m models.K8sEvent
			if err := m.MarkProcessedByID(event.ID, true); err != nil {
				klog.Errorf("标记事件已处理失败: %v", err)
			}
			continue
		}
		grouped[event.Cluster] = append(grouped[event.Cluster], event)
	}

	if len(grouped) == 0 {
		return nil
	}

	// 按集群批量推送
	for cluster, events := range grouped {
		if err := w.pushWebhookBatch(cluster, events); err != nil {
			klog.Errorf("批量Webhook推送失败: 集群=%s 错误=%v", cluster, err)
			// 失败则为该集群内所有事件增加重试次数
			for _, e := range events {
				if err := modelEvent.IncrementAttemptsByID(e.ID); err != nil {
					klog.Errorf("增加重试次数失败: %v", err)
				}
			}
		} else {
			// 推送成功则标记该集群内事件为已处理
			var m models.K8sEvent
			for _, e := range events {
				if err := m.MarkProcessedByID(e.ID, true); err != nil {
					klog.Errorf("标记事件已处理失败: %v", err)
				}
			}
		}
	}

	return nil
}

// pushWebhookBatch 推送批量事件的Webhook通知（按集群）
func (w *EventWorker) pushWebhookBatch(cluster string, events []*models.K8sEvent) error {
	// 按集群获取WebhookID列表
	webhookIDs, ok := w.cfg.Webhooks[cluster]
	if !ok || len(webhookIDs) == 0 {
		klog.V(6).Infof("集群 %s 未配置Webhook，跳过推送", cluster)
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

	// 生成批量摘要与原始JSON数组
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Event Warning 批量事件\n集群：[%s]\n数量：%d\n\n", cluster, len(events)))
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

	klog.V(6).Infof("批量Webhook推送成功: 集群=%s 事件数=%d", cluster, len(events))
	return nil
}
