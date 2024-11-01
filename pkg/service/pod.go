package service

import (
	"context"
	"io"

	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
)

type PodService struct {
}

func (p *PodService) StreamPodLogs(ctx context.Context, ns, name string, logOptions *v1.PodLogOptions) (io.ReadCloser, error) {

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
