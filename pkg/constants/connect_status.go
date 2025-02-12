package constants

// ClusterConnectStatus 集群连接状态
type ClusterConnectStatus string

const (
	ClusterConnectStatusConnected    ClusterConnectStatus = "connected"    // 已连接
	ClusterConnectStatusDisconnected ClusterConnectStatus = "disconnected" // 未连接
	ClusterConnectStatusFailed       ClusterConnectStatus = "failed"       // 连接失败
	ClusterConnectStatusConnecting   ClusterConnectStatus = "connecting"   // 连接中
)
