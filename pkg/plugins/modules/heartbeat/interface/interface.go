package heartbeatinterface

// HeartbeatManager 心跳管理器接口
type HeartbeatManager interface {
	// StartHeartbeat 启动心跳任务
	StartHeartbeat(clusterID string)
	// StopHeartbeat 停止心跳任务
	StopHeartbeat(clusterID string)
	// StartReconnect 启动自动重连
	StartReconnect(clusterID string)
	// StopReconnect 停止自动重连
	StopReconnect(clusterID string)
}

// 全局心跳管理器实例
var GlobalHeartbeatManager HeartbeatManager
