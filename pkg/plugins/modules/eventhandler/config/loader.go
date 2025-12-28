package config

import (
	"github.com/weibaohui/k8m/internal/dao"
	coremodels "github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/plugins/modules/eventhandler/models"
	"k8s.io/klog/v2"
)

// LoadAllFromDB 中文函数注释：
// 1. 读取所有启用的 K8sEventConfig 记录；
// 2. Worker/Watcher 等运行参数从平台全局配置中加载；
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

	var platformCfg coremodels.Config
	if err := dao.DB().First(&platformCfg).Error; err != nil {
		klog.V(6).Infof("读取平台全局配置失败，使用默认事件转发参数。错误: %v", err)
	}
	cfg := &EventHandlerConfig{
		Enabled: platformCfg.EventForwardEnabled || platformCfg.ID == 0,
		Watcher: WatcherConfig{
			BufferSize: defaultInt(platformCfg.EventWatcherBufferSize, 1000),
		},
		Worker: WorkerConfig{
			BatchSize:       defaultInt(platformCfg.EventWorkerBatchSize, 50),
			ProcessInterval: defaultInt(platformCfg.EventWorkerProcessInterval, 10),
			MaxRetries:      defaultInt(platformCfg.EventWorkerMaxRetries, 3),
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

