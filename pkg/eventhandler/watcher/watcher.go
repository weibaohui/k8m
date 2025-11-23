// Package watcher 实现Kubernetes事件监听器
package watcher

import (
	"context"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	utils2 "github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/eventhandler/model"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/kom"
	eventsv1 "k8s.io/api/events/v1"
	"k8s.io/apimachinery/pkg/watch"
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
	// 中文函数注释：使用定时任务每分钟检查所有已连接集群，未开启事件Watch则为其启动，并将告警事件入队处理
	klog.V(6).Infof("开始监听Kubernetes事件")

	ctx := utils2.GetContextWithAdmin()

	inst := cron.New()
	_, err := inst.AddFunc("@every 1m", func() {
		clusters := service.ClusterService().ConnectedClusters()
		for _, cluster := range clusters {
			if !cluster.GetClusterWatchStatus("event") {
				selectedCluster := service.ClusterService().ClusterID(cluster)

				var watcher watch.Interface
				var evt eventsv1.Event
				err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&evt).AllNamespace().Watch(&watcher).Error
				if err != nil {
					klog.Errorf("%s 创建Event监听器失败 %v", selectedCluster, err)
					continue
				}

				go func(clusterID string) {
					klog.V(6).Infof("%s 开始事件监听", clusterID)
					defer watcher.Stop()
					for e := range watcher.ResultChan() {
						if err := kom.Cluster(clusterID).WithContext(ctx).Tools().ConvertRuntimeObjectToTypedObject(e.Object, &evt); err != nil {
							klog.V(6).Infof("%s 无法将对象转换为 *events.v1.Event 类型: %v", clusterID, err)
							return
						}

						m := &model.Event{
							Type:   evt.Type,
							Reason: evt.Reason,
							Level: func() string {
								if evt.Type == "Warning" {
									return "warning"
								}
								return "normal"
							}(),
							Namespace: evt.Regarding.Namespace,
							Name:      evt.Regarding.Name,
							Message:   evt.Note,
							Timestamp: evt.EventTime.Time,
							Processed: false,
							Attempts:  0,
						}
						if m.EvtKey == "" {
							if string(evt.UID) != "" {
								m.EvtKey = string(evt.UID)
							} else {
								m.EvtKey = model.GenerateEvtKey(m.Namespace, "Event", m.Name, m.Reason, m.Message)
							}
						}

						if err := w.HandleEvent(m); err != nil {
							klog.V(6).Infof("%s 事件处理失败: %v", clusterID, err)
						} else {
							klog.V(6).Infof("%s 入队事件 [ %s/%s ] 类型=%s 原因=%s", clusterID, m.Namespace, m.Name, m.Type, m.Reason)
						}
					}
				}(selectedCluster)

				cluster.SetClusterWatchStarted("event", watcher)
			}
		}
	})
	if err != nil {
		klog.Errorf("新增Event状态定时更新任务报错: %v\n", err)
	}
	inst.Start()
	klog.V(6).Infof("新增Event状态定时更新任务【@every 1m】\n")

	<-w.ctx.Done()
	inst.Stop()
	return nil
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
				// 落库事件（统一使用模型方法）
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
				if err := ke.UpsertByEvtKey(); err != nil {
					klog.Errorf("存储/更新事件失败: %v", err)
				} else {
					klog.V(6).Infof("事件存储成功: %s", event.EvtKey)
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
	var m models.K8sEvent
	existing, err := m.GetByEvtKey(event.EvtKey)
	if err != nil {
		return fmt.Errorf("查询事件失败: %w", err)
	}
	if existing != nil {
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
