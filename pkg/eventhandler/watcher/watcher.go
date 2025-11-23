// Package watcher 实现Kubernetes事件监听器
package watcher

import (
	"context"
	"fmt"
	"time"

	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/eventhandler/model"
	"github.com/weibaohui/k8m/pkg/models"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// EventWatcher 事件监听器
type EventWatcher struct {
	client       kubernetes.Interface
	config       *model.EventHandlerConfig
	ruleMatcher  *RuleMatcher
	eventCh      chan *model.Event
	ctx          context.Context
	cancel       context.CancelFunc
	resyncPeriod time.Duration
}

// NewEventWatcher 创建事件监听器
func NewEventWatcher(client kubernetes.Interface, config *model.EventHandlerConfig) *EventWatcher {
	ctx, cancel := context.WithCancel(context.Background())

	return &EventWatcher{
		client:       client,
		config:       config,
		ruleMatcher:  NewRuleMatcher(&config.RuleConfig),
		eventCh:      make(chan *model.Event, config.Watcher.BufferSize),
		ctx:          ctx,
		cancel:       cancel,
		resyncPeriod: time.Duration(config.Watcher.ResyncInterval) * time.Second,
	}
}

// Start 启动事件监听器
func (w *EventWatcher) Start() error {
	if !w.config.Watcher.Enabled {
		klog.V(6).Infof("事件监听器未启用")
		return nil
	}

	klog.V(6).Infof("启动事件监听器")

	// 启动事件处理goroutine
	go w.processEvents()

	// 启动事件监听
	go w.watchEvents()

	return nil
}

// Stop 停止事件监听器
func (w *EventWatcher) Stop() {
	klog.V(6).Infof("停止事件监听器")
	w.cancel()
	close(w.eventCh)
}

// watchEvents 监听Kubernetes事件
func (w *EventWatcher) watchEvents() {
	for {
		select {
		case <-w.ctx.Done():
			return
		default:
			if err := w.doWatch(); err != nil {
				klog.Errorf("事件监听失败: %v", err)
				time.Sleep(5 * time.Second) // 失败后等待5秒重试
			}
		}
	}
}

// doWatch 执行事件监听
func (w *EventWatcher) doWatch() error {
	// 这里应该使用实际的Kubernetes事件监听
	// 由于需要集成到现有项目，我们先创建一个模拟的事件源
	klog.V(6).Infof("开始监听Kubernetes事件")

	// 模拟事件监听循环
	ticker := time.NewTicker(w.resyncPeriod)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return nil
		case <-ticker.C:
			// 这里应该调用实际的Kubernetes API获取事件
			// 暂时跳过，等集成时再实现
			continue
		}
	}
}

// processEvents 处理接收到的事件
func (w *EventWatcher) processEvents() {
	for {
		select {
		case <-w.ctx.Done():
			return
		case event, ok := <-w.eventCh:
			if !ok {
				return
			}

			// 初步过滤事件
			if w.shouldProcessEvent(event) {
				// 落库事件（统一使用 dao.DB()）
				ke := &models.K8sEvent{
					EvtKey:    event.EvtKey,
					Type:      event.Type,
					Reason:    event.Reason,
					Level:     event.Level,
					Namespace: event.Namespace,
					Name:      event.Name,
					Message:   event.Message,
					Timestamp: event.Timestamp,
					Processed: event.Processed,
					Attempts:  event.Attempts,
				}

				// 先查是否存在，避免重复
				var existing models.K8sEvent
				err := dao.DB().Where("evt_key = ?", ke.EvtKey).First(&existing).Error
				if err == nil {
					// 已存在，更新时间和消息
					if uErr := dao.DB().Model(&models.K8sEvent{}).Where("evt_key = ?", ke.EvtKey).Updates(map[string]any{
						"timestamp": ke.Timestamp,
						"message":   ke.Message,
					}).Error; uErr != nil {
						klog.Errorf("更新事件失败: %v", uErr)
					} else {
						klog.V(6).Infof("事件更新成功: %s", event.EvtKey)
					}
				} else {
					// 不存在则创建
					if cErr := dao.DB().Create(ke).Error; cErr != nil {
						klog.Errorf("创建事件失败: %v", cErr)
					} else {
						klog.V(6).Infof("事件存储成功: %s", event.EvtKey)
					}
				}
			}
		}
	}
}

// shouldProcessEvent 判断是否应该处理事件
func (w *EventWatcher) shouldProcessEvent(event *model.Event) bool {
	// 只处理警告类型事件
	if !event.IsWarning() {
		return false
	}

	// 应用规则匹配
	return w.ruleMatcher.Match(event)
}

// HandleEvent 处理单个事件（供外部调用）
func (w *EventWatcher) HandleEvent(event *model.Event) error {
	if event == nil {
		return fmt.Errorf("事件不能为空")
	}

	// 设置事件键
	if event.EvtKey == "" {
		event.EvtKey = model.GenerateEvtKey(event.Namespace, "Event", event.Name, event.Reason, event.Message)
	}

	// 检查事件是否已经存在
	var existing models.K8sEvent
	err := dao.DB().Where("evt_key = ?", event.EvtKey).First(&existing).Error
	if err == nil {
		klog.V(6).Infof("事件已存在，跳过: %s", event.EvtKey)
		return nil
	}

	// 发送到事件通道
	select {
	case w.eventCh <- event:
		return nil
	case <-time.After(1 * time.Second):
		return fmt.Errorf("事件通道已满，发送超时")
	}
}
