package service

import (
	"fmt"

	"github.com/weibaohui/kom/kom"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type nodeService struct {
}

func (n *nodeService) SetIPUsage(item unstructured.Unstructured) unstructured.Unstructured {
	nodeName := item.GetName()
	total, used, available := kom.DefaultCluster().Name(nodeName).Ctl().Node().IPUsage()

	// 设置或追加 annotations
	addAnnotations(&item, map[string]string{
		"ip.usage.total":     fmt.Sprintf("%d", total),
		"ip.usage.used":      fmt.Sprintf("%d", used),
		"ip.usage.available": fmt.Sprintf("%d", available),
	})

	return item
}

// addAnnotations 添加或更新 annotations
func addAnnotations(item *unstructured.Unstructured, newAnnotations map[string]string) {
	// 获取现有的 annotations
	annotations := item.GetAnnotations()
	if annotations == nil {
		// 如果不存在，初始化一个 map
		annotations = make(map[string]string)
	}

	// 追加或覆盖新数据
	for key, value := range newAnnotations {
		annotations[key] = value
	}

	// 设置回对象
	item.SetAnnotations(annotations)
}
