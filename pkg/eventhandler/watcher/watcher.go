// Package watcher 实现Kubernetes事件监听器
package watcher

import (
	"context"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	utils2 "github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/eventhandler/config"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/kom"
	eventsv1 "k8s.io/api/events/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog/v2"
)

// EventWatcher 事件监听器
type EventWatcher struct {
	cfg          *config.EventHandlerConfig
	ruleMatcher  *RuleMatcher
	eventCh      chan *models.K8sEvent
	ctx          context.Context
	cancel       context.CancelFunc
	resyncPeriod time.Duration
}

// NewEventWatcher 创建事件监听器
func NewEventWatcher() *EventWatcher {
	ctx, cancel := context.WithCancel(context.Background())
	cfg := config.DefaultEventHandlerConfig()
	return &EventWatcher{
		cfg:         cfg,
		ruleMatcher: NewRuleMatcher(&cfg.RuleConfig),
		eventCh:     make(chan *models.K8sEvent, cfg.Watcher.BufferSize),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Start 启动事件监听器
func (w *EventWatcher) Start() {
	if w.cfg.Enabled {
		klog.V(6).Infof("启动事件监听器")
		// 启动事件处理goroutine
		go w.processEvents()
		// 启动事件监听
		go w.watchEvents()
	} else {
		klog.V(6).Infof("事件转发功能未开启")
	}

}

// Stop 停止事件监听器
func (w *EventWatcher) Stop() {
	if w.cfg.Enabled {
		klog.V(6).Infof("停止事件监听器")
		w.cancel()
		close(w.eventCh)
	}

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

	inst := cron.New()
	_, err := inst.AddFunc("@every 1m", func() {
		clusters := service.ClusterService().ConnectedClusters()
		for _, cluster := range clusters {
			if !cluster.GetClusterWatchStatus("event") {
				selectedCluster := service.ClusterService().ClusterID(cluster)
				watcher := w.watchSingleCluster(selectedCluster)
				if watcher != nil {
					cluster.SetClusterWatchStarted("event", watcher)
				}
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

// watchSingleCluster 启动单个集群的事件监听
func (w *EventWatcher) watchSingleCluster(selectedCluster string) watch.Interface {
	ctx := utils2.GetContextWithAdmin()

	var watcher watch.Interface
	var evt eventsv1.Event
	err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&evt).AllNamespace().Watch(&watcher).Error
	if err != nil {
		klog.Errorf("%s 创建Event监听器失败 %v", selectedCluster, err)
		return nil
	}

	go func() {
		klog.V(6).Infof("%s 开始事件监听", selectedCluster)
		defer watcher.Stop()
		for e := range watcher.ResultChan() {
			if err := kom.Cluster(selectedCluster).WithContext(ctx).Tools().ConvertRuntimeObjectToTypedObject(e.Object, &evt); err != nil {
				klog.V(6).Infof("%s 无法将对象转换为 *events.v1.Event 类型: %v", selectedCluster, err)
				return
			}

			m := &models.K8sEvent{
				Type:      evt.Type,
				Reason:    evt.Reason,
				Cluster:   selectedCluster,
				Level:     evt.Type,
				Namespace: evt.Regarding.Namespace,
				Name:      evt.Regarding.Name,
				Message:   evt.Note,
				Timestamp: evt.EventTime.Time,
				Processed: false,
				Attempts:  0,
				EvtKey:    string(evt.UID),
			}

			if err := w.HandleEvent(m); err != nil {
				klog.V(6).Infof("%s 事件处理失败: %v", selectedCluster, err)
			} else {
				klog.V(6).Infof("%s 入队事件 [ %s/%s ] 类型=%s 原因=%s", selectedCluster, m.Namespace, m.Name, m.Type, m.Reason)
			}
		}
	}()

	return watcher
}

// HandleEvent 处理单个事件（供外部调用）
func (w *EventWatcher) HandleEvent(event *models.K8sEvent) error {
	if event == nil {
		return fmt.Errorf("事件不能为空")
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
func (w *EventWatcher) shouldProcessEvent(event *models.K8sEvent) bool {
	// 只处理警告类型事件
	if !event.IsWarning() {
		return false
	}

	// 应用规则匹配
	return w.ruleMatcher.Match(event)
}
