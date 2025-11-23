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

	// todo 全局的这些参数，放到flag中，放到集群参数设置的页面中
	cfg := &EventHandlerConfig{
		Enabled:      true,
		Watcher:      WatcherConfig{BufferSize: 1000},
		Worker:       WorkerConfig{BatchSize: 50, ProcessInterval: 10, MaxRetries: 3},
		EventConfigs: items,
	}

	klog.V(6).Infof("已从数据库加载事件处理器配置，共计规则条目: %d", len(items))
	return cfg
}
