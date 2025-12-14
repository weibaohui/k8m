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
	"github.com/weibaohui/k8m/pkg/service"
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
// 中文函数注释：根据当前配置的处理周期动态运行批处理任务，
// 当配置更新后会在下一次tick后动态调整定时器间隔，无需重启。
func (w *EventWorker) processLoop() {
	defer w.wg.Done()

	// 初始化定时器
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
				klog.Errorf("处理事件批次失败: %v", err)
			}
			// 动态调整定时器间隔
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
		var namespaces, names, reasons []string
		if ec.RuleNamespaces != "" {
			_ = json.Unmarshal([]byte(ec.RuleNamespaces), &namespaces)
		}
		if ec.RuleNames != "" { // 目前不参与匹配，但保留解析
			_ = json.Unmarshal([]byte(ec.RuleNames), &names)
		}

		if ec.RuleReasons != "" {
			_ = json.Unmarshal([]byte(ec.RuleReasons), &reasons)
		}

		rule := config.RuleConfig{Namespaces: namespaces, Names: names, Reasons: reasons, Reverse: ec.RuleReverse}

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
			if err := w.pushWebhookBatchForIDs(cluster, webhookIDs, events, ec.Name, ec.AIEnabled, ec.AIPromptTemplate); err != nil {
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
func (w *EventWorker) pushWebhookBatchForIDs(cluster string, webhookIDs []string, events []*models.K8sEvent, ruleName string, aiEnabled bool, aiTemplate string) error {
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
	sb.WriteString(fmt.Sprintf("Event Warning 事件\n规则：[%s]\n集群：[%s]\n数量：%d\n\n", ruleName, cluster, len(events)))
	for _, e := range events {
		sb.WriteString(fmt.Sprintf("资源：%s/%s\n类型：%s\n原因：%s\n消息：%s\n时间：%s\n\n",
			e.Namespace, e.Name, e.Type, e.Reason, e.Message, e.Timestamp.Format("2006-01-02 15:04:05")))
	}
	summary := sb.String()
	resultRaw := utils.ToJSONCompact(events)

	// AI总结：启用且事件数量>0时尝试；失败则回退并在结尾追加【AI总结失败】
	if aiEnabled && len(events) > 0 {
		if service.AIService().IsEnabled() {
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

			aiSummary, err := service.ChatService().ChatWithCtxNoHistory(w.ctx, prompt)
			if err != nil {
				klog.Errorf("AI总结失败，回退到字符串拼接: %v", err)
				summary = summary + "【AI总结失败】"
			} else {
				summary = aiSummary
			}
		} else {
			klog.V(6).Infof("AI服务未启用，跳过AI总结")
		}
	} else {
		if !aiEnabled {
			klog.V(6).Infof("规则AI总结未开启，跳过AI总结")
		} else if len(events) == 0 {
			klog.V(6).Infof("事件数量为0，跳过AI总结")
		}
	}

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
