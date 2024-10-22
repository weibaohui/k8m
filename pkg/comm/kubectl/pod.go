package kubectl

import (
	"context"
	"io"

	"github.com/weibaohui/kom/kom/poder"
	v1 "k8s.io/api/core/v1"
)

func (k8s *Kubectl) StreamPodLogs(ctx context.Context, ns, name string, logOptions *v1.PodLogOptions) (io.ReadCloser, error) {

	// 检查logOptions
	//  at most one of `sinceTime` or `sinceSeconds` may be specified
	if (logOptions.SinceTime != nil) && (logOptions.SinceSeconds != nil && *logOptions.SinceSeconds > 0) {
		// 同时设置，保留SinceSeconds
		logOptions.SinceTime = nil
	}
	if logOptions.SinceSeconds != nil && *logOptions.SinceSeconds == 0 {
		logOptions.SinceSeconds = nil
	}

	podLogs := poder.Instance().WithContext(ctx).Namespace(ns).Name(name).GetLogs(name, logOptions)
	logStream, err := podLogs.Stream(ctx)

	return logStream, err
}
