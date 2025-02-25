package service

import (
	"context"
	"io"
	"sync"

	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
)

type podService struct {
	// 存储Pod标签的map，key为集群ID，value为该集群下所有Pod的标签map
	podLabels map[string][]*PodLabels
	CountList []*StatusCount
	lock      sync.RWMutex
}

func (p *podService) StreamPodLogs(ctx context.Context, selectedCluster string, ns, name string, logOptions *v1.PodLogOptions) (io.ReadCloser, error) {

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
	err := kom.Cluster(selectedCluster).WithContext(ctx).
		Namespace(ns).Name(name).Ctl().Pod().
		ContainerName(logOptions.Container).GetLogs(&stream, logOptions).Error

	return stream, err
}
