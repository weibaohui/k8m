package service

import (
	"github.com/weibaohui/k8m/pkg/k8sgpt/analysis"
	"k8s.io/klog/v2"
)

func (c *clusterService) SetNodeStatusAggregated(selectedCluster string, true bool) {
	clusterConfig := c.GetClusterByID(selectedCluster)
	if clusterConfig == nil {
		return
	}
	clusterConfig.NodeStatusAggregated = true
}

// SetPodStatusAggregated 设置Pod状态聚合
func (c *clusterService) SetPodStatusAggregated(selectedCluster string, true bool) {
	clusterConfig := c.GetClusterByID(selectedCluster)
	if clusterConfig == nil {
		return
	}
	clusterConfig.PodStatusAggregated = true
}

// GetPodStatusAggregated 获取指定集群的Pod聚合状态
func (c *clusterService) GetPodStatusAggregated(selectedCluster string) bool {
	clusterConfig := c.GetClusterByID(selectedCluster)
	if clusterConfig == nil {
		return false
	}
	klog.V(6).Infof("获取Pod聚合状态: %s/%s: %v", clusterConfig.FileName, clusterConfig.ContextName, clusterConfig.PodStatusAggregated)
	return clusterConfig.PodStatusAggregated
}

// GetNodeStatusAggregated 获取指定集群的Node聚合状态
func (c *clusterService) GetNodeStatusAggregated(selectedCluster string) bool {
	clusterConfig := c.GetClusterByID(selectedCluster)
	if clusterConfig == nil {
		return false
	}
	klog.V(6).Infof("获取节点聚合状态: %s/%s: %v", clusterConfig.FileName, clusterConfig.ContextName, clusterConfig.NodeStatusAggregated)
	return clusterConfig.NodeStatusAggregated
}

// SetPVCStatusAggregated 设置指定集群的StorageClass聚合状态
func (c *clusterService) SetPVCStatusAggregated(selectedCluster string, true bool) {
	clusterConfig := c.GetClusterByID(selectedCluster)
	if clusterConfig == nil {
		return
	}
	klog.V(6).Infof("设置PVC存储类聚合状态: %s/%s: %v", clusterConfig.FileName, clusterConfig.ContextName, true)
	clusterConfig.PVCStatusAggregated = true
}
func (c *clusterService) GetPVCStatusAggregated(selectedCluster string) bool {
	clusterConfig := c.GetClusterByID(selectedCluster)
	if clusterConfig == nil {
		return false
	}
	return clusterConfig.PVCStatusAggregated
}

// SetPVStatusAggregated 设置指定集群的StorageClass聚合状态
func (c *clusterService) SetPVStatusAggregated(selectedCluster string, true bool) {
	clusterConfig := c.GetClusterByID(selectedCluster)
	if clusterConfig == nil {
		return
	}
	klog.V(6).Infof("设置PV存储类聚合状态: %s/%s: %v", clusterConfig.FileName, clusterConfig.ContextName, true)
	clusterConfig.PVStatusAggregated = true
}
func (c *clusterService) GetPVStatusAggregated(selectedCluster string) bool {
	clusterConfig := c.GetClusterByID(selectedCluster)
	if clusterConfig == nil {
		return false
	}
	return clusterConfig.PVStatusAggregated
}

// SetIngressStatusAggregated 设置Pod状态聚合
func (c *clusterService) SetIngressStatusAggregated(selectedCluster string, true bool) {
	clusterConfig := c.GetClusterByID(selectedCluster)
	if clusterConfig == nil {
		return
	}
	clusterConfig.IngressStatusAggregated = true
}
func (c *clusterService) GetIngressStatusAggregated(selectedCluster string) bool {
	clusterConfig := c.GetClusterByID(selectedCluster)
	if clusterConfig == nil {
		return false
	}
	return clusterConfig.IngressStatusAggregated
}

func (c *clusterService) SetClusterScanStatus(selectedCluster string, result *analysis.ResultWithStatus) {
	clusterConfig := c.GetClusterByID(selectedCluster)
	if clusterConfig == nil {
		return
	}

	clusterConfig.SetClusterScanStatus(result)

}

func (c *clusterService) GetClusterScanResult(selectedCluster string) *analysis.ResultWithStatus {
	clusterConfig := c.GetClusterByID(selectedCluster)
	if clusterConfig == nil {
		return nil
	}
	return clusterConfig.GetClusterScanResult()
}
