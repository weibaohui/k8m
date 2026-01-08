package watcher

import (
	"context"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	utils2 "github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/plugins/modules/eventhandler/config"
	"github.com/weibaohui/k8m/pkg/plugins/modules/eventhandler/models"
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/kom"
	eventsv1 "k8s.io/api/events/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/klog/v2"
)

// EventWatcher 中文函数注释：事件监听器。
type EventWatcher struct {
	cfg     *config.EventHandlerConfig
	eventCh chan *models.K8sEvent
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewEventWatcher 中文函数注释：创建事件监听器。
func NewEventWatcher() *EventWatcher {
	ctx, cancel := context.WithCancel(context.Background())
	cfg := config.DefaultEventHandlerConfig()
	return &EventWatcher{
		cfg:     cfg,
		eventCh: make(chan *models.K8sEvent, cfg.Watcher.BufferSize),
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Start 中文函数注释：启动事件监听器。
func (w *EventWatcher) Start() {
	if w.cfg.Enabled {
		klog.V(6).Infof("启动事件监听器")
		go w.processEvents()
		go w.watchEvents()
	} else {
		klog.V(6).Infof("事件转发功能未开启")
	}
}

// Stop 中文函数注释：无论当前配置开关状态如何，均立即停止监听器上下文。
func (w *EventWatcher) Stop() {
	if w == nil {
		return
	}
	klog.V(6).Infof("停止事件监听器")
	if w.cancel != nil {
		w.cancel()
	}
}

// watchEvents 中文函数注释：持续监听各集群事件，失败后重试。
func (w *EventWatcher) watchEvents() {
	for {
		select {
		case <-w.ctx.Done():
			return
		default:
			if err := w.doWatch(); err != nil {
				klog.V(6).Infof("事件监听失败: %v", err)
				time.Sleep(5 * time.Second)
			}
		}
	}
}

// doWatch 中文函数注释：使用定时任务每分钟检查所有已连接集群，未开启事件Watch则为其启动，并将告警事件入队处理。
func (w *EventWatcher) doWatch() error {
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
		klog.V(6).Infof("新增Event状态定时更新任务失败: %v", err)
	}
	inst.Start()
	klog.V(6).Infof("新增Event状态定时更新任务【@every 1m】")

	<-w.ctx.Done()
	inst.Stop()
	return nil
}

// watchSingleCluster 中文函数注释：启动单个集群的事件监听。
func (w *EventWatcher) watchSingleCluster(selectedCluster string) watch.Interface {
	ctx := utils2.GetContextWithAdminFromCtx(w.ctx)
	var watcher watch.Interface
	var evt eventsv1.Event
	err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&evt).AllNamespace().Watch(&watcher).Error
	if err != nil {
		klog.V(6).Infof("%s 创建Event监听器失败: %v", selectedCluster, err)
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
				Timestamp: func() time.Time {
					if !evt.EventTime.IsZero() {
						return evt.EventTime.Time
					}
					if !evt.ObjectMeta.CreationTimestamp.IsZero() {
						return evt.ObjectMeta.CreationTimestamp.Time
					}
					return time.Now()
				}(),
				Processed: false,
				Attempts:  0,
				EvtKey:    string(evt.UID),
			}

			if err := w.HandleEvent(m); err != nil {
				klog.V(6).Infof("%s 事件处理失败: %v", selectedCluster, err)
			} else {
				klog.V(6).Infof("%s 入队事件 [%s/%s] 类型=%s 原因=%s", selectedCluster, m.Namespace, m.Name, m.Type, m.Reason)
			}
		}
	}()

	return watcher
}

// HandleEvent 中文函数注释：处理单个事件（供外部调用）。
func (w *EventWatcher) HandleEvent(event *models.K8sEvent) error {
	if event == nil {
		return fmt.Errorf("事件不能为空")
	}
	if !w.shouldProcessEvent(event) {
		klog.V(6).Infof("事件 %s 不满足规则，跳过", event.EvtKey)
		return nil
	}

	timer := time.NewTimer(1 * time.Second)
	defer timer.Stop()
	for {
		select {
		case <-w.ctx.Done():
			return fmt.Errorf("监听器已停止，取消事件发送: %s", event.EvtKey)
		default:
		}

		select {
		case w.eventCh <- event:
			return nil
		default:
			klog.V(6).Infof("事件通道繁忙，等待重试发送: %s", event.EvtKey)
			select {
			case <-timer.C:
			case <-w.ctx.Done():
				if !timer.Stop() {
					<-timer.C
				}
				return fmt.Errorf("监听器已停止，取消事件发送: %s", event.EvtKey)
			}
		}
	}
}

// processEvents 中文函数注释：持久化接收到的事件。
func (w *EventWatcher) processEvents() {
	for {
		select {
		case <-w.ctx.Done():
			return
		case event, ok := <-w.eventCh:
			if !ok {
				return
			}
			if err := event.SaveEvent(); err != nil {
				klog.V(6).Infof("存储/更新事件失败: %v", err)
			} else {
				klog.V(6).Infof("事件存储成功: %s", event.EvtKey)
			}
		}
	}
}

// shouldProcessEvent 中文函数注释：判断是否应该处理事件。
func (w *EventWatcher) shouldProcessEvent(event *models.K8sEvent) bool {
	return event.IsWarning()
}
