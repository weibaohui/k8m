package service

import (
	"fmt"

	"github.com/weibaohui/kom/utils"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
)

type storageClassService struct {
}

// SetPVCCount 设置 PVC 数量
func (n *storageClassService) SetPVCCount(selectedCluster string, item unstructured.Unstructured) unstructured.Unstructured {
	name := item.GetName()
	// 从PVCService中获取PVC数量
	count := PVCService().GetPVCCount(selectedCluster, name)
	klog.V(6).Infof("SetPVCCount: %s/%s, count: %d", selectedCluster, name, count)
	utils.AddOrUpdateAnnotations(&item, map[string]string{
		"pvc.count": fmt.Sprintf("%d", count),
	})

	return item
}

// SetPVCount 设置 PV 数量
func (n *storageClassService) SetPVCount(selectedCluster string, item unstructured.Unstructured) unstructured.Unstructured {
	name := item.GetName()
	// 从PVService中获取PV数量
	count := PVService().GetPVCount(selectedCluster, name)
	klog.V(6).Infof("SetPVCount: %s/%s, count: %d", selectedCluster, name, count)
	utils.AddOrUpdateAnnotations(&item, map[string]string{
		"pv.count": fmt.Sprintf("%d", count),
	})

	return item
}
