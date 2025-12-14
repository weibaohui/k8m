package eventhandler

import (
	"sync"

	"github.com/robfig/cron/v3"
	"github.com/weibaohui/k8m/pkg/eventhandler/watcher"
	"github.com/weibaohui/k8m/pkg/eventhandler/worker"
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
)

// StartEventForwarding 启动事件监听与处理（受平台开关控制）
// 中文函数注释：读取平台配置，仅在开启总开关时启动 Watcher 与 Worker；若已运行则跳过。
func StartEventForwarding() error {
	cfg, err := service.ConfigService().GetConfig()
	if err != nil || cfg == nil {
		klog.V(6).Infof("读取平台配置失败，事件转发未启动：%v", err)
		return err
	}
	if !cfg.EventForwardEnabled {
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
	lastSnapshot.watcherBuffer = cfg.EventWatcherBufferSize
	lastSnapshot.batchSize = cfg.EventWorkerBatchSize
	lastSnapshot.intervalSec = cfg.EventWorkerProcessInterval
	lastSnapshot.maxRetries = cfg.EventWorkerMaxRetries
	klog.V(6).Infof("事件监听与处理已启动")
	return nil
}

// StopEventForwarding 停止事件监听与处理
// 中文函数注释：停止当前运行的 Watcher 与 Worker，并清理内部引用。
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

// SyncEventForwardingFromConfig 按最新平台配置同步事件转发状态与参数
// 中文函数注释：每次调用均读取数据库最新配置；若开关或参数变化，则执行启停或更新，保持与平台配置一致。
func SyncEventForwardingFromConfig() {
	cfg, err := service.ConfigService().GetConfig()
	if err != nil || cfg == nil {
		klog.V(6).Infof("读取平台配置失败，跳过事件转发同步：%v", err)
		return
	}
	// 开关变化：直接启停
	lock.Lock()
	enabledSnapshot := lastSnapshot.enabled
	lock.Unlock()
	if cfg.EventForwardEnabled != enabledSnapshot {
		if cfg.EventForwardEnabled {
			_ = StartEventForwarding()
		} else {
			StopEventForwarding()
		}
		return
	}
	// 参数变化：在开启状态下更新
	if cfg.EventForwardEnabled {
		changed := cfg.EventWatcherBufferSize != lastSnapshot.watcherBuffer ||
			cfg.EventWorkerBatchSize != lastSnapshot.batchSize ||
			cfg.EventWorkerProcessInterval != lastSnapshot.intervalSec ||
			cfg.EventWorkerMaxRetries != lastSnapshot.maxRetries
		if changed {
			lock.Lock()
			// Watcher 无法动态更新缓存大小，需重启
			if currentWatch != nil {
				currentWatch.Stop()
			}
			currentWatch = watcher.NewEventWatcher()
			currentWatch.Start()
			// Worker 支持动态更新
			if currentWork != nil {
				currentWork.UpdateConfig()
			} else {
				currentWork = worker.NewEventWorker()
				currentWork.Start()
			}

			lastSnapshot.enabled = true
			lastSnapshot.watcherBuffer = cfg.EventWatcherBufferSize
			lastSnapshot.batchSize = cfg.EventWorkerBatchSize
			lastSnapshot.intervalSec = cfg.EventWorkerProcessInterval
			lastSnapshot.maxRetries = cfg.EventWorkerMaxRetries
			lock.Unlock()
			klog.V(6).Infof("已按最新平台配置同步事件转发参数并生效")
		}
	}
}

// StartEventForwardingWatch 启动事件转发配置监听
// 中文函数注释：设置一个定时器，后台不断更新事件转发配置，保持与平台配置一致。
// 动作：每 1 分钟调用一次 SyncEventForwardingFromConfig 函数，同步事件转发配置。
// 启停或更新：根据平台配置开关状态，启动或停止事件监听与处理；若开关或参数变化，更新配置。
func StartEventForwardingWatch() {
	// 设置一个定时器，后台不断更新事件转发配置
	cronLock.Lock()
	defer cronLock.Unlock()
	if eventForwardCron != nil {
		klog.V(6).Infof("事件转发配置定时任务已在运行，跳过重复启动")
		return
	}
	eventForwardCron = cron.New()
	_, err := eventForwardCron.AddFunc("@every 1m", func() {
		// 延迟启动cron
		SyncEventForwardingFromConfig()
	})
	if err != nil {
		klog.Errorf("新增事件转发配置定时任务报错: %v\n", err)
	}
	eventForwardCron.Start()
	klog.V(6).Infof("新增事件转发配置定时任务【@every 1m】\n")
}

// StopEventForwardingWatch 停止事件转发配置监听
// 中文函数注释：优雅停止定时任务，避免重复任务或资源泄漏。
func StopEventForwardingWatch() {
	StopEventForwarding()
	cronLock.Lock()
	defer cronLock.Unlock()
	if eventForwardCron != nil {
		eventForwardCron.Stop()
		eventForwardCron = nil
		klog.V(6).Infof("事件转发配置定时任务已停止")
	} else {
		klog.V(6).Infof("事件转发配置定时任务未运行")
	}
}
