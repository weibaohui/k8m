package service

import (
	"fmt"

	"github.com/weibaohui/kom/utils"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
)

type ingressClassService struct {
}

// SetPVCCount 设置 PVC 数量
func (n *ingressClassService) SetIngressCount(selectedCluster string, item unstructured.Unstructured) unstructured.Unstructured {
	name := item.GetName()
	// 从PVCService中获取PVC数量
	count := IngressService().GetIngressCount(selectedCluster, name)
	klog.V(6).Infof("SetIngressCount: %s/%s, count: %d", selectedCluster, name, count)
	utils.AddOrUpdateAnnotations(&item, map[string]string{
		"ingress.count": fmt.Sprintf("%d", count),
	})

	return item
}
