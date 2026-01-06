package service

import (
	"encoding/json"
	"net/http"

	"github.com/weibaohui/k8m/pkg/flag"
)

// HeartbeatConfig 心跳配置结构
type HeartbeatConfig struct {
	HeartbeatIntervalSeconds    int `json:"heartbeat_interval_seconds"`     // 心跳间隔时间（秒）
	HeartbeatFailureThreshold   int `json:"heartbeat_failure_threshold"`    // 心跳失败阈值
	ReconnectMaxIntervalSeconds int `json:"reconnect_max_interval_seconds"` // 重连最大间隔时间（秒）
	MaxRetryAttempts            int `json:"max_retry_attempts"`             // 最大重试次数
}

// GetHeartbeatConfig 获取当前心跳配置
func GetHeartbeatConfig(w http.ResponseWriter, r *http.Request) {
	cfg := flag.Init()
	config := HeartbeatConfig{
		HeartbeatIntervalSeconds:    cfg.HeartbeatIntervalSeconds,
		HeartbeatFailureThreshold:   cfg.HeartbeatFailureThreshold,
		ReconnectMaxIntervalSeconds: cfg.ReconnectMaxIntervalSeconds,
		MaxRetryAttempts:            cfg.MaxRetryAttempts,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code": 0,
		"msg":  "success",
		"data": config,
	})
}

// SaveHeartbeatConfig 保存心跳配置
func SaveHeartbeatConfig(w http.ResponseWriter, r *http.Request) {
	var config HeartbeatConfig
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 1,
			"msg":  "参数错误: " + err.Error(),
		})
		return
	}

	// 保存配置到全局标志变量
	cfg := flag.Init()
	cfg.HeartbeatIntervalSeconds = config.HeartbeatIntervalSeconds
	cfg.HeartbeatFailureThreshold = config.HeartbeatFailureThreshold
	cfg.ReconnectMaxIntervalSeconds = config.ReconnectMaxIntervalSeconds
	cfg.MaxRetryAttempts = config.MaxRetryAttempts

	// 这里应该添加保存到配置文件的逻辑

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code": 0,
		"msg":  "保存成功",
	})
}
