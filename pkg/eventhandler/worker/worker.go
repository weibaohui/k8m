// Package worker 实现事件处理Worker
package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/weibaohui/k8m/pkg/eventhandler/model"
	"github.com/weibaohui/k8m/pkg/eventhandler/store"
	"k8s.io/klog/v2"
)

// EventWorker 事件处理Worker
type EventWorker struct {
	store        store.EventStore
	config       *model.EventHandlerConfig
	ctx          context.Context
	cancel       context.CancelFunc
	wg           sync.WaitGroup
	processMutex sync.Mutex
}

// NewEventWorker 创建事件处理Worker
func NewEventWorker(store store.EventStore, config *model.EventHandlerConfig) *EventWorker {
	ctx, cancel := context.WithCancel(context.Background())

	return &EventWorker{
		store:  store,
		config: config,
		ctx:    ctx,
		cancel: cancel,
	}
}

// Start 启动Worker
func (w *EventWorker) Start() error {
	if !w.config.Worker.Enabled {
		klog.V(6).Infof("事件处理Worker未启用")
		return nil
	}

	klog.V(6).Infof("启动事件处理Worker")

	w.wg.Add(1)
	go w.processLoop()

	return nil
}

// Stop 停止Worker
func (w *EventWorker) Stop() {
	klog.V(6).Infof("停止事件处理Worker")
	w.cancel()
	w.wg.Wait()
}

// processLoop 处理循环
func (w *EventWorker) processLoop() {
	defer w.wg.Done()

	ticker := time.NewTicker(time.Duration(w.config.Worker.ProcessInterval) * time.Millisecond)
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

	// 获取未处理的事件
	events, err := w.store.GetUnprocessed(w.ctx, w.config.Worker.BatchSize)
	if err != nil {
		return fmt.Errorf("获取未处理事件失败: %w", err)
	}

	if len(events) == 0 {
		return nil
	}

	klog.V(6).Infof("开始处理事件批次: %d个事件", len(events))

	for _, event := range events {
		if err := w.processEvent(event); err != nil {
			klog.Errorf("处理事件失败: %v", err)
			// 增加重试次数
			if err := w.store.IncrementAttempts(w.ctx, event.ID); err != nil {
				klog.Errorf("增加重试次数失败: %v", err)
			}
		}
	}

	return nil
}

// processEvent 处理单个事件
func (w *EventWorker) processEvent(event *model.Event) error {
	klog.V(6).Infof("处理事件: %s", event.EvtKey)

	// 检查重试次数
	if event.Attempts >= w.config.Worker.MaxRetries {
		klog.Warningf("事件达到最大重试次数，标记为已处理: %s", event.EvtKey)
		return w.store.UpdateProcessed(w.ctx, event.ID, true)
	}

	// 应用二次过滤（聚合、去重、限流等）
	if w.shouldFilterEvent(event) {
		klog.V(6).Infof("事件被过滤: %s", event.EvtKey)
		return w.store.UpdateProcessed(w.ctx, event.ID, true)
	}

	// 推送Webhook
	// TODO: 集成webhook推送功能
	klog.V(6).Infof("事件处理完成，准备推送: %s", event.EvtKey)

	// 标记为已处理
	return w.store.UpdateProcessed(w.ctx, event.ID, true)
}

// shouldFilterEvent 判断是否应该过滤事件
func (w *EventWorker) shouldFilterEvent(event *model.Event) bool {
	// TODO: 实现更复杂的过滤逻辑
	// 1. 聚合规则：同一资源的相似事件可以聚合
	// 2. 限流规则：防止同一事件频繁推送
	// 3. 去重规则：避免重复推送相同事件

	// 简单的示例：如果事件消息包含特定关键词，则过滤
	filterKeywords := []string{"test", "debug"}
	for _, keyword := range filterKeywords {
		if contains(event.Message, keyword) {
			return true
		}
	}

	return false
}

// contains 检查字符串是否包含子字符串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsSubstring(s, substr)))
}

// containsSubstring 检查字符串是否包含子字符串（内部实现）
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
