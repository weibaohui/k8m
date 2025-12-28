package config

import (
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/plugins/modules/eventhandler/models"
	"k8s.io/klog/v2"
)

// LoadAllFromDB 中文函数注释：
// 1. 读取所有启用的 K8sEventConfig 记录；
// 2. Worker/Watcher 等运行参数从插件配置表中加载；
// 3. 若不存在启用规则，则返回 nil 交由上层使用默认配置。
func LoadAllFromDB() *EventHandlerConfig {
	var items []models.K8sEventConfig
	if err := dao.DB().Where("enabled = ?", true).Order("id asc").Find(&items).Error; err != nil {
		klog.V(6).Infof("事件处理器配置未从数据库加载，使用默认配置。错误: %v", err)
		return nil
	}
	if len(items) == 0 {
		klog.V(6).Infof("数据库中不存在启用的事件处理器配置，使用默认配置")
		return nil
	}

	setting, err := models.GetOrCreateEventForwardSetting()
	if err != nil || setting == nil {
		klog.V(6).Infof("读取事件转发插件配置失败，使用默认事件转发参数。错误: %v", err)
		setting = models.DefaultEventForwardSetting()
	}
	cfg := &EventHandlerConfig{
		Enabled: setting.EventForwardEnabled,
		Watcher: WatcherConfig{
			BufferSize: defaultInt(setting.EventWatcherBufferSize, 1000),
		},
		Worker: WorkerConfig{
			BatchSize:       defaultInt(setting.EventWorkerBatchSize, 50),
			ProcessInterval: defaultInt(setting.EventWorkerProcessInterval, 10),
			MaxRetries:      defaultInt(setting.EventWorkerMaxRetries, 3),
		},
		EventConfigs: items,
	}

	klog.V(6).Infof("已从数据库加载事件处理器配置，共计规则条目: %d", len(items))
	return cfg
}

// defaultInt 中文函数注释：当传入的数值为0或小于0时，返回提供的默认值。
func defaultInt(v int, def int) int {
	if v <= 0 {
		return def
	}
	return v
}
