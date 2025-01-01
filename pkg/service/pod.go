package service

import (
	"context"
	"fmt"
	"io"

	"github.com/weibaohui/kom/kom"
	"github.com/weibaohui/kom/utils"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type podService struct {
}

func (p *podService) StreamPodLogs(ctx context.Context, ns, name string, logOptions *v1.PodLogOptions) (io.ReadCloser, error) {

	// 检查logOptions
	//  at most one of `sinceTime` or `sinceSeconds` may be specified
	if (logOptions.SinceTime != nil) && (logOptions.SinceSeconds != nil && *logOptions.SinceSeconds > 0) {
		// 同时设置，保留SinceSeconds
		logOptions.SinceTime = nil
	}
	if logOptions.SinceSeconds != nil && *logOptions.SinceSeconds == 0 {
		logOptions.SinceSeconds = nil
	}
	var stream io.ReadCloser
	err := kom.DefaultCluster().WithContext(ctx).Namespace(ns).Name(name).ContainerName(logOptions.Container).GetLogs(&stream, logOptions).Error

	return stream, err
}

// SetAllocatedStatus 设置节点的分配状态
func (p *podService) SetAllocatedStatus(item unstructured.Unstructured) unstructured.Unstructured {
	podName := item.GetName()
	table := kom.DefaultCluster().Name(podName).Namespace(item.GetNamespace()).Resource(&v1.Pod{}).Ctl().Pod().ResourceUsageTable()
	for _, row := range table {
		if row.ResourceType == "cpu" {
			// 设置或追加 annotations
			utils.AddOrUpdateAnnotations(&item, map[string]string{
				"cpu.request":         fmt.Sprintf("%s", row.Request),
				"cpu.requestFraction": fmt.Sprintf("%s", row.RequestFraction),
				"cpu.limit":           fmt.Sprintf("%s", row.Limit),
				"cpu.limitFraction":   fmt.Sprintf("%s", row.LimitFraction),
			})
		} else if row.ResourceType == "memory" {
			// 设置或追加 annotations
			utils.AddOrUpdateAnnotations(&item, map[string]string{
				"memory.request":         fmt.Sprintf("%s", row.Request),
				"memory.requestFraction": fmt.Sprintf("%s", row.RequestFraction),
				"memory.limit":           fmt.Sprintf("%s", row.Limit),
				"memory.limitFraction":   fmt.Sprintf("%s", row.LimitFraction),
			})
		}
	}

	return item
}
