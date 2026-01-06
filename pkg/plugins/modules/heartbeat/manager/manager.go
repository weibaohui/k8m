package manager

import "github.com/weibaohui/k8m/pkg/plugins/modules/heartbeat/service"

// 全局心跳管理器实例
var globalHeartbeatManager *service.HeartbeatManager

// GetHeartbeatManager 获取全局心跳管理器实例
func GetHeartbeatManager() *service.HeartbeatManager {
	return globalHeartbeatManager
}

// SetHeartbeatManager 设置全局心跳管理器实例
func SetHeartbeatManager(manager *service.HeartbeatManager) {
	globalHeartbeatManager = manager
}
