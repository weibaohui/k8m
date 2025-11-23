package config

import (
    "encoding/json"

    "github.com/weibaohui/k8m/internal/dao"
    "github.com/weibaohui/k8m/pkg/models"
    "k8s.io/klog/v2"
)

// LoadFromDB 从数据库加载事件处理器配置，并转换为可用的配置结构
// 中文函数注释：读取启用的事件处理器配置（最新一条），将JSON格式的规则字段解析为切片/映射。
// 若数据库未配置或解析失败，打印中文日志并返回nil，由调用方使用内置默认值。
func LoadFromDB() *EventHandlerConfig {
    var item models.K8sEventConfig
    err := dao.DB().Where("enabled = ?", true).Order("id desc").Limit(1).First(&item).Error
    if err != nil {
        klog.V(6).Infof("事件处理器配置未从数据库加载，使用默认配置。错误: %v", err)
        return nil
    }

    var namespaces []string
    var labels map[string]string
    var reasons []string
    var types []string

    if item.RuleNamespaces != "" {
        if err := json.Unmarshal([]byte(item.RuleNamespaces), &namespaces); err != nil {
            klog.V(6).Infof("解析规则命名空间失败，将使用空列表: %v", err)
            namespaces = nil
        }
    }
    if item.RuleLabels != "" {
        if err := json.Unmarshal([]byte(item.RuleLabels), &labels); err != nil {
            klog.V(6).Infof("解析规则标签失败，将使用空映射: %v", err)
            labels = nil
        }
    }
    if item.RuleReasons != "" {
        if err := json.Unmarshal([]byte(item.RuleReasons), &reasons); err != nil {
            klog.V(6).Infof("解析规则原因失败，将使用空列表: %v", err)
            reasons = nil
        }
    }
    if item.RuleTypes != "" {
        if err := json.Unmarshal([]byte(item.RuleTypes), &types); err != nil {
            klog.V(6).Infof("解析规则类型失败，将使用空列表: %v", err)
            types = nil
        }
    }

    cfg := &EventHandlerConfig{
        Enabled: item.Enabled,
        Watcher: WatcherConfig{
            BufferSize: item.WatcherBufferSize,
        },
        Worker: WorkerConfig{
            BatchSize:       item.WorkerBatchSize,
            ProcessInterval: item.WorkerProcessInterval,
            MaxRetries:      item.WorkerMaxRetries,
        },
        RuleConfig: RuleConfig{
            Namespaces: namespaces,
            Labels:     labels,
            Reasons:    reasons,
            Types:      types,
            Reverse:    item.RuleReverse,
        },
    }

    klog.V(6).Infof("已从数据库加载事件处理器配置: Enabled=%v, BufferSize=%d, BatchSize=%d", cfg.Enabled, cfg.Watcher.BufferSize, cfg.Worker.BatchSize)
    return cfg
}

