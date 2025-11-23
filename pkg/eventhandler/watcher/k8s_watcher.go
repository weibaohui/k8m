// Package watcher 实现Kubernetes事件监听器集成
package watcher

import (
	"fmt"
	"time"

	"github.com/weibaohui/k8m/pkg/eventhandler/model"
	"github.com/weibaohui/k8m/pkg/eventhandler/store"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

// K8sEventWatcher Kubernetes事件监听器
type K8sEventWatcher struct {
	watcher *EventWatcher
}

// NewK8sEventWatcher 创建Kubernetes事件监听器
func NewK8sEventWatcher(client kubernetes.Interface, store store.EventStore, config *model.EventHandlerConfig) *K8sEventWatcher {
	return &K8sEventWatcher{
		watcher: NewEventWatcher(client, store, config),
	}
}

// Start 启动监听器
func (k *K8sEventWatcher) Start() error {
	return k.watcher.Start()
}

// Stop 停止监听器
func (k *K8sEventWatcher) Stop() {
	k.watcher.Stop()
}

// HandleK8sEvent 处理Kubernetes事件
func (k *K8sEventWatcher) HandleK8sEvent(eventType watch.EventType, event *corev1.Event) error {
	if event == nil {
		return fmt.Errorf("事件不能为空")
	}

	// 转换Kubernetes事件为我们的模型
	evt := &model.Event{
		Type:      event.Type,
		Reason:    event.Reason,
		Level:     getEventLevel(event),
		Namespace: event.Namespace,
		Name:      event.Name,
		Message:   event.Message,
		Timestamp: event.LastTimestamp.Time,
		Processed: false,
		Attempts:  0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	return k.watcher.HandleEvent(evt)
}

// getEventLevel 获取事件级别
func getEventLevel(event *corev1.Event) string {
	if event.Type == "Warning" {
		return "warning"
	}
	return "normal"
}
