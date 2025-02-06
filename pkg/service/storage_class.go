package service

import (
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/utils"
	v1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
)

// storageClassStatusTTL storage class状态缓存时间
// 要跟watch中的定时处理器保持一致
var storageClassStatusTTL = 5 * time.Minute

type storageClassService struct {
}

// SetPVCCount 设置 PVC 数量
func (n *storageClassService) SetPVCCount(selectedCluster string, item unstructured.Unstructured) unstructured.Unstructured {
	name := item.GetName()
	count, err := n.CachePVCCount(selectedCluster, name)
	if err != nil {
		return item
	}
	klog.V(4).Infof("SetPVCCount: %s/%s, count: %d", selectedCluster, name, count)
	utils.AddOrUpdateAnnotations(&item, map[string]string{
		"pvc.count": fmt.Sprintf("%d", count),
	})

	return item
}
func (n *storageClassService) SyncStorageClassStatus(selectedCluster string) {
	klog.V(6).Infof("Sync StorageClass Status")
	var storageClasses []v1.StorageClass
	err := kom.Cluster(selectedCluster).Resource(&v1.StorageClass{}).WithCache(storageClassStatusTTL).List(&storageClasses).Error
	if err != nil {
		klog.Errorf("监听StorageClass失败:%v", err)
	}
	for _, storageClass := range storageClasses {
		_, _ = n.CachePVCCount(selectedCluster, storageClass.Name)
	}
	ClusterService().SetStorageClassStatusAggregated(selectedCluster, true)
}
func (n *storageClassService) Watch() {
	// 设置一个定时器，后台不断更新storageClass状态
	inst := cron.New()
	_, err := inst.AddFunc("@every 1m", func() {
		// 延迟启动cron
		clusters := ClusterService().ConnectedClusters()
		for _, cluster := range clusters {
			selectedCluster := ClusterService().ClusterID(cluster)
			n.SyncStorageClassStatus(selectedCluster)
			klog.V(6).Infof("执行定时更新StorageClass状态%s", selectedCluster)
		}
	})
	if err != nil {
		klog.Errorf("Error add cron job for storageClass: %v\n", err)
	}
	inst.Start()
	klog.V(6).Infof("新增StorageClass状态定时更新任务【@every 5m】\n")
}

func (n *storageClassService) CachePVCCount(selectedCluster string, storageClassName string) (int, error) {
	cacheKey := fmt.Sprintf("%s/%s", "StorageClassPVCCount", storageClassName)
	return utils.GetOrSetCache(kom.Cluster(selectedCluster).ClusterCache(), cacheKey, storageClassStatusTTL, func() (int, error) {
		return kom.Cluster(selectedCluster).Name(storageClassName).WithCache(storageClassStatusTTL).Ctl().StorageClass().PVCCount()
	})
}

func (n *storageClassService) RemoveStorageClassStatusCache(selectedCluster string, storageClassName string) {
	StorageClassPVCCountKey := fmt.Sprintf("%s/%s", "StorageClassPVCCount", storageClassName)
	kom.Cluster(selectedCluster).ClusterCache().Del(StorageClassPVCCountKey)
}
