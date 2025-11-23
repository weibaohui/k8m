package config

import (
	"encoding/json"
	"strings"

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

	clusterRules := make(map[string]RuleConfig)
	clusterWebhooks := make(map[string][]string)
	// 解析每条记录的规则并分配到集群
	// 如果有多个重叠的配置，只会保留最后一个，前面的会被覆盖
	for _, it := range items {
		var namespaces []string
		var names []string
		var labels []string
		var reasons []string
		var types []string

		//关键字列表，多个用逗号分隔
		if it.RuleNamespaces != "" {
			if err := json.Unmarshal([]byte(it.RuleNamespaces), &namespaces); err != nil {
				klog.V(6).Infof("解析规则命名空间失败，将使用空列表: %v", err)
				namespaces = nil
			}
		}
		//关键字列表，多个用逗号分隔
		if it.RuleNames != "" {
			if err := json.Unmarshal([]byte(it.RuleNames), &names); err != nil {
				klog.V(6).Infof("解析规则命名失败，将使用空列表: %v", err)
				names = nil
			}
		}
		//关键字列表，多个用逗号分隔，app=k8m,env=dev
		if it.RuleLabels != "" {
			if err := json.Unmarshal([]byte(it.RuleLabels), &labels); err != nil {
				klog.V(6).Infof("解析规则标签失败，将使用空列表: %v", err)
				labels = nil
			}
		}
		//关键字列表，多个用逗号分隔
		if it.RuleReasons != "" {
			if err := json.Unmarshal([]byte(it.RuleReasons), &reasons); err != nil {
				klog.V(6).Infof("解析规则原因失败，将使用空列表: %v", err)
				reasons = nil
			}
		}

		// 关键字列表，多个用逗号分隔
		if it.RuleTypes != "" {
			if err := json.Unmarshal([]byte(it.RuleTypes), &types); err != nil {
				klog.V(6).Infof("解析规则类型失败，将使用空列表: %v", err)
				types = nil
			}
		}

		rule := RuleConfig{
			Namespaces: namespaces,
			Labels:     labels,
			Reasons:    reasons,
			Types:      types,
			Reverse:    it.RuleReverse,
		}

		// 按集群分配
		for _, c := range strings.Split(it.Clusters, ",") {
			cluster := strings.TrimSpace(c)
			if cluster == "" {
				continue
			}
			// 如果有多个重叠的配置，只会保留最后一个，前面的会被覆盖
			clusterRules[cluster] = rule
		}

		// 按集群分配WebhookID
		for _, c := range strings.Split(it.Clusters, ",") {
			cluster := strings.TrimSpace(c)
			if cluster == "" {
				continue
			}
			clusterWebhooks[cluster] = append(clusterWebhooks[cluster], strings.Split(it.Webhooks, ",")...)
		}
	}

	cfg := &EventHandlerConfig{
		Enabled:      true,
		Watcher:      WatcherConfig{BufferSize: 1000},
		Worker:       WorkerConfig{BatchSize: 50, ProcessInterval: 1, MaxRetries: 3},
		ClusterRules: clusterRules,
		Webhooks:     clusterWebhooks,
	}

	klog.V(6).Infof("已从数据库加载事件处理器配置，共计规则条目: %d", len(clusterRules))
	return cfg
}
