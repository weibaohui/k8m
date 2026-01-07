package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/plugins"
	"github.com/weibaohui/k8m/pkg/plugins/modules"
	"github.com/weibaohui/k8m/pkg/plugins/modules/ai/service"
	"github.com/weibaohui/k8m/pkg/plugins/modules/eventhandler/config"
	"github.com/weibaohui/k8m/pkg/plugins/modules/eventhandler/models"

	"github.com/weibaohui/k8m/pkg/plugins/modules/webhook"
	"k8s.io/klog/v2"
)

// EventWorker 中文函数注释：事件处理Worker。
type EventWorker struct {
	cfg          *config.EventHandlerConfig
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	processMutex sync.Mutex
}

var defaultWorker *EventWorker

// NewEventWorker 中文函数注释：创建事件处理Worker。
func NewEventWorker() *EventWorker {
	ctx, cancel := context.WithCancel(context.Background())
	cfg := config.DefaultEventHandlerConfig()
	ew := &EventWorker{
		cfg:    cfg,
		ctx:    ctx,
		cancel: cancel,
	}
	defaultWorker = ew
	return ew
}

// Start 中文函数注释：启动Worker。
func (w *EventWorker) Start() {
	if w.cfg.Enabled {
		klog.V(6).Infof("启动事件处理Worker")
		w.wg.Add(1)
		go w.processLoop()
	} else {
		klog.V(6).Infof("事件转发功能未开启")
	}
}

// Stop 中文函数注释：无论当前配置开关状态如何，均立即停止事件处理循环并等待退出。
func (w *EventWorker) Stop() {
	if w == nil {
		return
	}
	klog.V(6).Infof("停止事件处理Worker")
	if w.cancel != nil {
		w.cancel()
	}
	w.wg.Wait()
}

// processLoop 中文函数注释：根据当前配置的处理周期动态运行批处理任务。
func (w *EventWorker) processLoop() {
	defer w.wg.Done()

	w.processMutex.Lock()
	interval := w.cfg.Worker.ProcessInterval
	w.processMutex.Unlock()
	if interval <= 0 {
		interval = 10
	}
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			if err := w.processBatch(); err != nil {
				klog.V(6).Infof("处理事件批次失败: %v", err)
			}
			w.processMutex.Lock()
			newInterval := w.cfg.Worker.ProcessInterval
			w.processMutex.Unlock()
			if newInterval <= 0 {
				newInterval = 10
			}
			if newInterval != interval {
				ticker.Stop()
				interval = newInterval
				ticker = time.NewTicker(time.Duration(interval) * time.Second)
				klog.V(6).Infof("事件处理器批次处理周期已更新为: %d 秒", interval)
			}
		}
	}
}

// UpdateConfig 中文函数注释：动态刷新事件处理配置（从数据库重新加载）。
func (w *EventWorker) UpdateConfig() {
	if w == nil {
		return
	}
	newCfg := config.DefaultEventHandlerConfig()
	if newCfg == nil {
		return
	}
	w.processMutex.Lock()
	w.cfg = newCfg
	w.processMutex.Unlock()
	klog.V(6).Infof("事件处理配置已更新，立即生效")
}

// Instance 中文函数注释：获取全局事件处理Worker实例。
func Instance() *EventWorker {
	return defaultWorker
}

// processBatch 中文函数注释：按批次获取未处理事件，逐条按每条事件配置进行过滤与推送。
func (w *EventWorker) processBatch() error {
	w.processMutex.Lock()
	defer w.processMutex.Unlock()

	var modelEvent models.K8sEvent
	k8sEvents, err := modelEvent.ListUnprocessed(w.cfg.Worker.BatchSize)
	if err != nil {
		return fmt.Errorf("获取未处理事件失败: %w", err)
	}
	if len(k8sEvents) == 0 {
		return nil
	}
	klog.V(6).Infof("开始处理事件批次: %d个事件", len(k8sEvents))

	processedIDs := make(map[int64]bool)
	matchedIDs := make(map[int64]bool)

	for _, ec := range w.cfg.EventConfigs {
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

		var namespaces, names, reasons []string
		if ec.RuleNamespaces != "" {
			_ = json.Unmarshal([]byte(ec.RuleNamespaces), &namespaces)
		}
		if ec.RuleNames != "" {
			_ = json.Unmarshal([]byte(ec.RuleNames), &names)
		}
		if ec.RuleReasons != "" {
			_ = json.Unmarshal([]byte(ec.RuleReasons), &reasons)
		}
		rule := config.RuleConfig{Namespaces: namespaces, Names: names, Reasons: reasons, Reverse: ec.RuleReverse}

		grouped := make(map[string][]*models.K8sEvent)
		for _, event := range k8sEvents {
			if processedIDs[event.ID] {
				continue
			}
			if event.Attempts >= w.cfg.Worker.MaxRetries {
				klog.V(6).Infof("事件达到最大重试次数，标记为已处理: %s", event.EvtKey)
				var m models.K8sEvent
				if err := m.MarkProcessedByID(event.ID, true); err != nil {
					klog.V(6).Infof("标记事件已处理失败: %v", err)
				}
				processedIDs[event.ID] = true
				continue
			}
			if _, ok := clusters[event.Cluster]; !ok {
				continue
			}
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

		for cluster, events := range grouped {
			if err := w.pushWebhookBatchForIDs(cluster, webhookIDs, events, ec.Name, ec.AIEnabled, ec.AIPromptTemplate); err != nil {
				klog.V(6).Infof("批量Webhook推送失败: 规则=%s 集群=%s 错误=%v", ec.Name, cluster, err)
				for _, e := range events {
					if err := modelEvent.IncrementAttemptsByID(e.ID); err != nil {
						klog.V(6).Infof("增加重试次数失败: %v", err)
					}
				}
			} else {
				var m models.K8sEvent
				for _, e := range events {
					if err := m.MarkProcessedByID(e.ID, true); err != nil {
						klog.V(6).Infof("标记事件已处理失败: %v", err)
					} else {
						processedIDs[e.ID] = true
					}
				}
			}
		}
	}

	for _, event := range k8sEvents {
		if !processedIDs[event.ID] && !matchedIDs[event.ID] {
			var m models.K8sEvent
			if err := m.MarkProcessedByID(event.ID, true); err != nil {
				klog.V(6).Infof("标记事件已处理失败: %v", err)
			}
		}
	}
	return nil
}

// pushWebhookBatchForIDs 中文函数注释：根据指定的 webhookID 列表批量推送事件。
func (w *EventWorker) pushWebhookBatchForIDs(cluster string, webhookIDs []string, events []*models.K8sEvent, ruleName string, aiEnabled bool, aiTemplate string) error {
	if len(webhookIDs) == 0 {
		klog.V(6).Infof("规则 %s 未配置Webhook，跳过推送", ruleName)
		return nil
	}

	if len(webhookIDs) == 0 {
		klog.V(6).Infof("规则 %s 未找到可用的webhook接收器，跳过推送", ruleName)
		return nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Event Warning 事件\n规则：[%s]\n集群：[%s]\n数量：%d\n\n", ruleName, cluster, len(events)))
	for _, e := range events {
		sb.WriteString(fmt.Sprintf("资源：%s/%s\n类型：%s\n原因：%s\n消息：%s\n时间：%s\n\n",
			e.Namespace, e.Name, e.Type, e.Reason, e.Message, e.Timestamp.Format("2006-01-02 15:04:05")))
	}
	summary := sb.String()
	resultRaw := utils.ToJSONCompact(events)

	if aiEnabled && len(events) > 0 {

		if plugins.ManagerInstance().IsRunning(modules.PluginNameAI) {
			customTemplate := aiTemplate
			if strings.TrimSpace(customTemplate) == "" {
				customTemplate = `请先输出统计数据（含集群名称、规则名称、数量等基本信息）
再逐条列出关键错误信息
附加简单的分析
总体不超过300字`
			}
			prompt := `以下是k8s集群事件记录，请你进行总结。
基本要求：
1、仅做汇总，不要解释
2、不需要解决方案。
3、可以合理使用表情符号。

附加要求：
%s

以下是JSON格式的事件列表：
%s
`
			prompt = fmt.Sprintf(prompt, customTemplate, resultRaw)

			aiSummary, err := service.GetChatService().ChatWithCtxNoHistory(w.ctx, prompt)
			if err != nil {
				klog.V(6).Infof("AI总结失败，回退到字符串拼接: %v", err)
				summary = summary + "【AI总结失败】"
			} else {
				summary = aiSummary
			}
		} else {
			klog.V(6).Infof("AI服务未启用，跳过AI总结")
		}
	}

	results := webhook.PushMsgToAllTargetByIDs(summary, resultRaw, webhookIDs)

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
