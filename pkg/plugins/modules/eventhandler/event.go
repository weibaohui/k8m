package eventhandler

import (
	"context"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/weibaohui/k8m/pkg/plugins/modules/eventhandler/models"
	"github.com/weibaohui/k8m/pkg/plugins/modules/eventhandler/watcher"
	"github.com/weibaohui/k8m/pkg/plugins/modules/eventhandler/worker"
	"github.com/weibaohui/k8m/pkg/service"
	"k8s.io/klog/v2"
)

var (
	lock             sync.Mutex
	cronLock         sync.Mutex
	currentWatch     *watcher.EventWatcher
	currentWork      *worker.EventWorker
	eventForwardCron *cron.Cron
	lastSnapshot     struct {
		enabled       bool
		watcherBuffer int
		batchSize     int
		intervalSec   int
		maxRetries    int
	}

	leaderWatchMu     sync.Mutex
	leaderWatchCancel context.CancelFunc
)

// StartLeaderWatch 中文函数注释：启动主备状态监听，根据主节点状态启停事件转发。
func StartLeaderWatch() {
	leaderWatchMu.Lock()
	defer leaderWatchMu.Unlock()

	if leaderWatchCancel != nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	leaderWatchCancel = cancel

	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		last := false
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				isLeader := service.LeaderService().IsCurrentLeader()
				if isLeader && !last {
					klog.V(6).Infof("检测到当前实例成为Leader，启动事件转发")
					StartEventForwardingWatch()
				}
				if !isLeader && last {
					klog.V(6).Infof("检测到当前实例不再是Leader，停止事件转发")
					StopEventForwardingWatch()
				}
				last = isLeader
			}
		}
	}()
}

// StopLeaderWatch 中文函数注释：停止主备状态监听，并停止事件转发。
func StopLeaderWatch() {
	leaderWatchMu.Lock()
	cancel := leaderWatchCancel
	leaderWatchCancel = nil
	leaderWatchMu.Unlock()

	if cancel != nil {
		cancel()
	}
	StopEventForwardingWatch()
}

// StartEventForwarding 中文函数注释：读取平台配置，仅在开启总开关时启动 Watcher 与 Worker；若已运行则跳过。
func StartEventForwarding() error {
	setting, err := models.GetOrCreateEventForwardSetting()
	if err != nil || setting == nil {
		klog.V(6).Infof("读取事件转发插件配置失败，事件转发未启动：%v", err)
		return err
	}
	if !setting.EventForwardEnabled {
		klog.V(6).Infof("事件转发总开关关闭，未启动事件监听与处理")
		return nil
	}
	lock.Lock()
	defer lock.Unlock()
	if currentWatch == nil {
		currentWatch = watcher.NewEventWatcher()
		currentWatch.Start()
	}
	if currentWork == nil {
		currentWork = worker.NewEventWorker()
		currentWork.Start()
	} else {
		currentWork.UpdateConfig()
	}
	lastSnapshot.enabled = true
	lastSnapshot.watcherBuffer = setting.EventWatcherBufferSize
	lastSnapshot.batchSize = setting.EventWorkerBatchSize
	lastSnapshot.intervalSec = setting.EventWorkerProcessInterval
	lastSnapshot.maxRetries = setting.EventWorkerMaxRetries
	klog.V(6).Infof("事件监听与处理已启动")
	return nil
}

// StopEventForwarding 中文函数注释：停止当前运行的 Watcher 与 Worker，并清理内部引用。
func StopEventForwarding() {
	lock.Lock()
	defer lock.Unlock()
	if currentWork != nil {
		currentWork.Stop()
		currentWork = nil
	} else {
		klog.V(6).Infof("事件处理者为nil")
	}
	if currentWatch != nil {
		currentWatch.Stop()
		currentWatch = nil
	} else {
		klog.V(6).Infof("事件监听者为nil")
	}
	lastSnapshot.enabled = false
	klog.V(6).Infof("事件监听与处理已停止")
}

// SyncEventForwardingFromConfig 中文函数注释：每次调用均读取数据库最新配置；若开关或参数变化，则执行启停或更新，保持与平台配置一致。
func SyncEventForwardingFromConfig() {
	setting, err := models.GetOrCreateEventForwardSetting()
	if err != nil || setting == nil {
		klog.V(6).Infof("读取事件转发插件配置失败，跳过事件转发同步：%v", err)
		return
	}
	lock.Lock()
	enabledSnapshot := lastSnapshot.enabled
	lock.Unlock()
	if setting.EventForwardEnabled != enabledSnapshot {
		if setting.EventForwardEnabled {
			_ = StartEventForwarding()
		} else {
			StopEventForwarding()
		}
		return
	}
	if setting.EventForwardEnabled {
		changed := setting.EventWatcherBufferSize != lastSnapshot.watcherBuffer ||
			setting.EventWorkerBatchSize != lastSnapshot.batchSize ||
			setting.EventWorkerProcessInterval != lastSnapshot.intervalSec ||
			setting.EventWorkerMaxRetries != lastSnapshot.maxRetries
		if changed {
			lock.Lock()
			if currentWatch != nil {
				currentWatch.Stop()
			}
			currentWatch = watcher.NewEventWatcher()
			currentWatch.Start()
			if currentWork != nil {
				currentWork.UpdateConfig()
			} else {
				currentWork = worker.NewEventWorker()
				currentWork.Start()
			}

			lastSnapshot.enabled = true
			lastSnapshot.watcherBuffer = setting.EventWatcherBufferSize
			lastSnapshot.batchSize = setting.EventWorkerBatchSize
			lastSnapshot.intervalSec = setting.EventWorkerProcessInterval
			lastSnapshot.maxRetries = setting.EventWorkerMaxRetries
			lock.Unlock()
			klog.V(6).Infof("已按最新事件转发插件配置同步参数并生效")
		}
	}
}

// StartEventForwardingWatch 中文函数注释：设置一个定时器，后台不断更新事件转发配置，保持与平台配置一致。
func StartEventForwardingWatch() {
	if eventForwardCron != nil {
		klog.V(6).Infof("事件转发配置定时任务已在运行，跳过重复启动")
		return
	}
	cronLock.Lock()
	eventForwardCron = cron.New(
		cron.WithChain(
			cron.Recover(cron.DefaultLogger),
			cron.SkipIfStillRunning(cron.DefaultLogger),
		),
	)
	_, err := eventForwardCron.AddFunc("@every 1m", func() {
		SyncEventForwardingFromConfig()
	})
	cronLock.Unlock()

	if err != nil {
		klog.V(6).Infof("新增事件转发配置定时任务失败: %v", err)
	}
	eventForwardCron.Start()
	klog.V(6).Infof("新增事件转发配置定时任务【@every 1m】")
}

// StopEventForwardingWatch 中文函数注释：优雅停止定时任务，避免重复任务或资源泄漏。
func StopEventForwardingWatch() {
	cronLock.Lock()
	if eventForwardCron != nil {
		eventForwardCron.Stop()
		eventForwardCron = nil
		klog.V(6).Infof("事件转发配置定时任务已停止")
	} else {
		klog.V(6).Infof("事件转发配置定时任务未运行")
	}
	cronLock.Unlock()
	StopEventForwarding()
}
