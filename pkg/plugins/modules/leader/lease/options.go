package lease

import "time"

// Options 中文函数注释：Lease 管理器的初始化选项，包含命名空间、续约与时长设置等。
type Options struct {
	Namespace                 string
	LeaseDurationSeconds      int
	LeaseRenewIntervalSeconds int
	ResyncPeriod              time.Duration
	ClusterID                 string // ClusterID 指定的集群唯一ID（文件名/Context），优先使用该集群的配置
}

