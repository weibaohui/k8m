package config

import (
	"github.com/weibaohui/k8m/internal/dao"
	"github.com/weibaohui/k8m/pkg/models"
	"k8s.io/klog/v2"
)

// LoadAllFromDB 从数据库加载所有启用的事件处理器配置（多条），并按集群组装规则
// 中文函数注释：
// 1. 读取所有启用的 K8sEventConfig 记录；
// 2. 将每条记录的规则字段(JSON)解析为 RuleConfig；
// 3. 根据记录中的 Clusters 字段（逗号分隔）分配到对应集群的规则；
// 4. Worker/Watcher 等运行参数取最近一条记录的值（如不存在则走默认）。
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

	// 从平台全局配置中加载事件转发参数
	var platformCfg models.Config
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

// defaultInt 返回有效值或默认值
// 中文函数注释：当传入的数值为0或小于0时，返回提供的默认值。
func defaultInt(v int, def int) int {
	if v <= 0 {
		return def
	}
	return v
}
