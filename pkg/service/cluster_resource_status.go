package service

import "k8s.io/klog/v2"

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

// SetStorageClassStatusAggregated 设置指定集群的StorageClass聚合状态
func (c *clusterService) SetStorageClassStatusAggregated(selectedCluster string, true bool) {
	clusterConfig := c.GetClusterByID(selectedCluster)
	if clusterConfig == nil {
		return
	}
	klog.V(6).Infof("设置存储类聚合状态: %s/%s: %v", clusterConfig.FileName, clusterConfig.ContextName, true)
	clusterConfig.StorageClassStatusAggregated = true
}
func (c *clusterService) GetStorageClassStatusAggregated(selectedCluster string) bool {
	clusterConfig := c.GetClusterByID(selectedCluster)
	if clusterConfig == nil {
		return false
	}
	return clusterConfig.StorageClassStatusAggregated
}
