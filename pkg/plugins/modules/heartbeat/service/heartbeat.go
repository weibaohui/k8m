package service

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/flag"
	"github.com/weibaohui/k8m/pkg/service"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

// HeartbeatManager 心跳管理服务
type HeartbeatManager struct {
	// 心跳管理
	heartbeatCancel           sync.Map // 心跳取消函数
	HeartbeatIntervalSeconds  int      // 心跳间隔秒数
	HeartbeatFailureThreshold int      // 心跳失败阈值

	// 自动重连管理
	reconnectCancel             sync.Map // 自动重连取消函数
	ReconnectMaxIntervalSeconds int      // 自动重连最大退避秒数
	MaxRetryAttempts            int      // 最大重试次数
}

// NewHeartbeatManager 创建心跳管理服务实例
func NewHeartbeatManager() *HeartbeatManager {
	cfg := flag.Init()
	hm := &HeartbeatManager{
		HeartbeatIntervalSeconds:    cfg.HeartbeatIntervalSeconds,
		HeartbeatFailureThreshold:   cfg.HeartbeatFailureThreshold,
		ReconnectMaxIntervalSeconds: cfg.ReconnectMaxIntervalSeconds,
		MaxRetryAttempts:            cfg.MaxRetryAttempts,
	}
	if hm.HeartbeatIntervalSeconds <= 0 {
		hm.HeartbeatIntervalSeconds = 30
	}
	if hm.HeartbeatFailureThreshold <= 0 {
		hm.HeartbeatFailureThreshold = 3
	}
	if hm.ReconnectMaxIntervalSeconds <= 0 {
		hm.ReconnectMaxIntervalSeconds = 3600
	}
	if hm.MaxRetryAttempts <= 0 {
		hm.MaxRetryAttempts = 100
	}
	return hm
}

// StartHeartbeat 启动心跳任务
func (h *HeartbeatManager) StartHeartbeat(clusterID string) {

	// 如果已有心跳，先停止
	if cancelInterface, ok := h.heartbeatCancel.Load(clusterID); ok {
		if cancel, ok := cancelInterface.(context.CancelFunc); ok && cancel != nil {
			cancel()
		}
		h.heartbeatCancel.Delete(clusterID)
	}

	cluster := service.ClusterService().GetClusterByID(clusterID)
	if cluster == nil {
		klog.V(6).Infof("启动心跳失败：未找到集群 %s", clusterID)
		return
	}
	// 仅在已连接时启动心跳
	if cluster.ClusterConnectStatus != constants.ClusterConnectStatusConnected {
		klog.V(6).Infof("集群 %s 非已连接状态，心跳不启动", clusterID)
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	h.heartbeatCancel.Store(clusterID, cancel)

	interval := time.Duration(h.HeartbeatIntervalSeconds) * time.Second
	ticker := time.NewTicker(interval)
	klog.V(6).Infof("集群 %s 心跳启动，间隔 %ds，失败阈值 %d", clusterID, h.HeartbeatIntervalSeconds, h.HeartbeatFailureThreshold)

	go func() {
		defer ticker.Stop()
		failureCount := 0
		for {
			select {
			case <-ctx.Done():
				klog.V(6).Infof("集群 %s 心跳已停止", clusterID)
				return
			case <-ticker.C:
				// 若集群不再是已连接状态，则停止心跳
				if cluster.ClusterConnectStatus != constants.ClusterConnectStatusConnected {
					klog.V(6).Infof("集群 %s 心跳检测：状态已非已连接，停止心跳", clusterID)
					cancel()
					return
				}
				// restConfig 必须存在
				if cluster.GetRestConfig() == nil {
					failureCount++
					klog.V(6).Infof("集群 %s 心跳检测失败：restConfig 不存在（累计失败 %d）", clusterID, failureCount)
					h.AppendHeartbeatRecord(cluster, false, time.Now())
				} else {
					clientset, err := kubernetes.NewForConfig(cluster.GetRestConfig())
					if err != nil {
						failureCount++
						klog.V(6).Infof("集群 %s 创建 clientset 失败：%v（累计失败 %d）", clusterID, err, failureCount)
						h.AppendHeartbeatRecord(cluster, false, time.Now())
					} else {
						sv, err := clientset.Discovery().ServerVersion()
						if err != nil {
							failureCount++
							klog.V(6).Infof("集群 %s 心跳检测读取版本失败：%v（累计失败 %d）", clusterID, err, failureCount)
							cluster.Err = err.Error()
							h.AppendHeartbeatRecord(cluster, false, time.Now())
						} else {
							failureCount = 0
							if sv != nil {
								cluster.ServerVersion = sv.GitVersion
							}
							klog.V(6).Infof("集群 %s 心跳检测成功，当前版本：%s", clusterID, cluster.ServerVersion)
							h.AppendHeartbeatRecord(cluster, true, time.Now())
						}
					}
				}

				if failureCount >= h.HeartbeatFailureThreshold {
					// 达到失败阈值，切换为断开并停止心跳，并执行重连
					cluster.ClusterConnectStatus = constants.ClusterConnectStatusDisconnected
					klog.V(6).Infof("集群 %s 心跳连续失败达到阈值，状态切换为未连接，启动自动重连", clusterID)

					// 停止当前心跳循环
					cancel()

					// 启动自动重连
					h.StartReconnect(clusterID)
					return
				}
			}
		}
	}()
}

// StopHeartbeat 停止指定集群的心跳任务
func (h *HeartbeatManager) StopHeartbeat(clusterID string) {
	if cancelInterface, ok := h.heartbeatCancel.Load(clusterID); ok {
		if cancel, ok := cancelInterface.(context.CancelFunc); ok && cancel != nil {
			cancel()
		}
		h.heartbeatCancel.Delete(clusterID)
		klog.V(6).Infof("集群 %s 心跳任务已停止", clusterID)
	}
}

// StartReconnect 启动指定集群的自动重连循环
func (h *HeartbeatManager) StartReconnect(clusterID string) {
	// 初始化重连配置默认值
	if h.ReconnectMaxIntervalSeconds <= 0 {
		h.ReconnectMaxIntervalSeconds = 3600
	}
	if h.MaxRetryAttempts <= 0 {
		h.MaxRetryAttempts = 100
	}

	// 若已有自动重连任务，先停止
	if cancelInterface, ok := h.reconnectCancel.Load(clusterID); ok {
		if cancel, ok := cancelInterface.(context.CancelFunc); ok && cancel != nil {
			cancel()
		}
		h.reconnectCancel.Delete(clusterID)
	}

	ctx, cancel := context.WithCancel(context.Background())
	h.reconnectCancel.Store(clusterID, cancel)

	klog.V(6).Infof("集群 %s 自动重连循环启动（最大退避 %ds，最大重试次数 %d）", clusterID, h.ReconnectMaxIntervalSeconds, h.MaxRetryAttempts)

	go func(id string) {
		attempt := 0
		backoff := 1 // 初始退避秒数
		maxIntervalSeconds := h.ReconnectMaxIntervalSeconds
		maxRetryAttempts := h.MaxRetryAttempts

		for {
			select {
			case <-ctx.Done():
				klog.V(6).Infof("集群 %s 自动重连循环已停止", id)
				return
			default:
			}

			// 若已连接则结束重连
			if service.ClusterService().IsConnected(id) {
				klog.V(6).Infof("集群 %s 已连接，自动重连循环结束", id)
				cancel()
				h.reconnectCancel.Delete(id)
				return
			}

			attempt++
			// 检查是否超过最大重试次数
			if attempt > maxRetryAttempts {
				klog.V(6).Infof("集群 %s 自动重连已达到最大重试次数 %d，停止重连", id, maxRetryAttempts)
				cancel()
				h.reconnectCancel.Delete(id)
				return
			}

			klog.V(6).Infof("集群 %s 自动重连第 %d 次尝试：先断开清理后重连", id, attempt)

			// 尝试连接
			service.ClusterService().Connect(id)

			// 若连接成功，结束重连循环
			if service.ClusterService().IsConnected(id) {
				klog.V(6).Infof("集群 %s 自动重连成功", id)
				cancel()
				h.reconnectCancel.Delete(id)
				return
			}

			// 指数退避，封顶
			backoff = min(backoff*2, maxIntervalSeconds)
			klog.V(6).Infof("集群 %s 自动重连失败，%ds 后重试", id, backoff)
			time.Sleep(time.Duration(backoff) * time.Second)
		}
	}(clusterID)
}

// StopReconnect 停止指定集群的自动重连循环
func (h *HeartbeatManager) StopReconnect(clusterID string) {
	if cancelInterface, ok := h.reconnectCancel.Load(clusterID); ok {
		if cancel, ok := cancelInterface.(context.CancelFunc); ok && cancel != nil {
			cancel()
		}
		h.reconnectCancel.Delete(clusterID)
		klog.V(6).Infof("集群 %s 自动重连循环已停止", clusterID)
	}
}

// UpdateSettings 更新心跳设置
func (h *HeartbeatManager) UpdateSettings() {
	cfg := flag.Init()
	h.HeartbeatIntervalSeconds = cfg.HeartbeatIntervalSeconds
	h.HeartbeatFailureThreshold = cfg.HeartbeatFailureThreshold
	h.ReconnectMaxIntervalSeconds = cfg.ReconnectMaxIntervalSeconds
	h.MaxRetryAttempts = cfg.MaxRetryAttempts

	klog.V(6).Infof("更新心跳设置：间隔 %d 秒，失败阈值 %d，最大重连间隔 %d 秒，最大重试次数 %d",
		h.HeartbeatIntervalSeconds, h.HeartbeatFailureThreshold, h.ReconnectMaxIntervalSeconds, h.MaxRetryAttempts)
}

// AppendHeartbeatRecord 追加一条心跳记录，并裁剪为阈值长度
func (h *HeartbeatManager) AppendHeartbeatRecord(cluster *service.ClusterConfig, success bool, ts time.Time) {
	if cluster == nil {
		return
	}
	if cluster.HeartbeatHistory == nil {
		cluster.HeartbeatHistory = make([]service.HeartbeatRecord, 0)
	}
	rec := service.HeartbeatRecord{
		Success: success,
		Time:    ts.Local().Format("2006-01-02 15:04:05"),
	}
	cluster.HeartbeatHistory = append(cluster.HeartbeatHistory, rec)
	threshold := h.HeartbeatFailureThreshold
	if threshold <= 0 {
		threshold = 3
	}
	if len(cluster.HeartbeatHistory) > threshold {
		cluster.HeartbeatHistory = cluster.HeartbeatHistory[len(cluster.HeartbeatHistory)-threshold:]
	}
	for i := range cluster.HeartbeatHistory {
		cluster.HeartbeatHistory[i].Index = i + 1
	}
}

// GetHeartbeatStatus 获取所有集群的心跳状态
func GetHeartbeatStatus(w http.ResponseWriter, r *http.Request) {
	clusters := service.ClusterService().AllClusters()
	statusList := make([]map[string]interface{}, 0, len(clusters))

	for _, cluster := range clusters {
		status := map[string]interface{}{
			"cluster_id":        cluster.ClusterID,
			"cluster_name":      cluster.ClusterName,
			"connect_status":    cluster.ClusterConnectStatus,
			"server_version":    cluster.ServerVersion,
			"last_heartbeat":    "",
			"heartbeat_history": cluster.HeartbeatHistory,
			"failure_count":     0,
		}

		// 获取最后一次心跳记录
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

		statusList = append(statusList, status)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code": 0,
		"msg":  "success",
		"data": statusList,
	})
}
