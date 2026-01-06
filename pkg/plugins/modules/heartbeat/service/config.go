package service

import (
	"encoding/json"
	"net/http"

	"github.com/weibaohui/k8m/pkg/flag"
	"github.com/weibaohui/k8m/pkg/service"
	"k8s.io/klog/v2"
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

	// 获取当前数据库配置
	dbConfig, err := service.ConfigService().GetConfig()
	if err != nil {
		klog.V(6).Infof("获取数据库配置失败: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 1,
			"msg":  "获取配置失败: " + err.Error(),
		})
		return
	}

	// 更新心跳相关配置
	dbConfig.HeartbeatIntervalSeconds = config.HeartbeatIntervalSeconds
	dbConfig.HeartbeatFailureThreshold = config.HeartbeatFailureThreshold
	dbConfig.ReconnectMaxIntervalSeconds = config.ReconnectMaxIntervalSeconds
	dbConfig.MaxRetryAttempts = config.MaxRetryAttempts

	// 保存到数据库
	if err := service.ConfigService().UpdateConfig(dbConfig); err != nil {
		klog.V(6).Infof("保存心跳配置失败: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 1,
			"msg":  "保存配置失败: " + err.Error(),
		})
		return
	}

	klog.V(6).Infof("心跳配置已保存: 间隔=%d秒, 失败阈值=%d, 最大重连间隔=%d秒, 最大重试次数=%d",
		config.HeartbeatIntervalSeconds,
		config.HeartbeatFailureThreshold,
		config.ReconnectMaxIntervalSeconds,
		config.MaxRetryAttempts)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code": 0,
		"msg":  "保存成功",
	})
}
