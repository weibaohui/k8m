package controller

import (
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/plugins/modules/heartbeat/models"
	hs "github.com/weibaohui/k8m/pkg/plugins/modules/heartbeat/service"
	"github.com/weibaohui/k8m/pkg/response"
	"github.com/weibaohui/k8m/pkg/service"
	"k8s.io/klog/v2"
)

type Controller struct {
}

// HeartbeatConfig 心跳配置结构
type HeartbeatConfig struct {
	HeartbeatIntervalSeconds    int `json:"heartbeat_interval_seconds"`     // 心跳间隔时间（秒）
	HeartbeatFailureThreshold   int `json:"heartbeat_failure_threshold"`    // 心跳失败阈值
	ReconnectMaxIntervalSeconds int `json:"reconnect_max_interval_seconds"` // 重连最大间隔时间（秒）
	MaxRetryAttempts            int `json:"max_retry_attempts"`             // 最大重试次数
}

// GetHeartbeatStatus 获取所有集群的心跳状态
func (h *Controller) GetHeartbeatStatus(c *response.Context) {
	clusters := service.ClusterService().AllClusters()
	statusList := make([]map[string]interface{}, 0, len(clusters))

	for _, cluster := range clusters {
		status := map[string]any{
			"cluster_id":        cluster.ClusterID,
			"cluster_name":      cluster.ClusterName,
			"connect_status":    cluster.ClusterConnectStatus,
			"server_version":    cluster.ServerVersion,
			"last_heartbeat":    "",
			"heartbeat_history": cluster.HeartbeatHistory,
			"failure_count":     0,
		}

		// 获取最后一次心跳记录
		cluster.HeartbeatMu.RLock()
		if len(cluster.HeartbeatHistory) > 0 {
			lastRecord := cluster.HeartbeatHistory[len(cluster.HeartbeatHistory)-1]
			status["last_heartbeat"] = lastRecord.Time

			// 计算连续失败次数
			failureCount := 0
			for i := len(cluster.HeartbeatHistory) - 1; i >= 0; i-- {
				if !cluster.HeartbeatHistory[i].Success {
					failureCount++
				} else {
					break
				}
			}
			status["failure_count"] = failureCount
		}
		cluster.HeartbeatMu.RUnlock()
		statusList = append(statusList, status)
	}

	amis.WriteJsonData(c, statusList)
}

// GetHeartbeatConfig 获取当前心跳配置
func (h *Controller) GetHeartbeatConfig(c *response.Context) {
	cfg, err := models.GetOrCreateHeartbeatSetting()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonData(c, cfg)
}

// SaveHeartbeatConfig 保存心跳配置
func (h *Controller) SaveHeartbeatConfig(c *response.Context) {
	var config HeartbeatConfig
	if err := c.ShouldBind(&config); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 构造心跳配置
	setting := &models.HeartbeatSetting{
		HeartbeatIntervalSeconds:    config.HeartbeatIntervalSeconds,
		HeartbeatFailureThreshold:   config.HeartbeatFailureThreshold,
		ReconnectMaxIntervalSeconds: config.ReconnectMaxIntervalSeconds,
		MaxRetryAttempts:            config.MaxRetryAttempts,
	}

	// 保存到数据库
	if _, err := models.UpdateHeartbeatSetting(setting); err != nil {
		klog.V(6).Infof("保存心跳配置失败: %v", err)
		amis.WriteJsonError(c, err)
		return
	}

	hs.NewHeartbeatManager().UpdateSettings()

	klog.V(6).Infof("心跳配置已保存: 间隔=%d秒, 失败阈值=%d, 最大重连间隔=%d秒, 最大重试次数=%d",
		config.HeartbeatIntervalSeconds,
		config.HeartbeatFailureThreshold,
		config.ReconnectMaxIntervalSeconds,
		config.MaxRetryAttempts)

	amis.WriteJsonOK(c)

}
