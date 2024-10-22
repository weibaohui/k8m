package kubectl

import (
	"context"
	"io"
	"sort"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	// json := &utils.JSONUtils{}
	// klog.V(2).Infof(json.ToJSON(logOptions))
	// 获取 Pod 日志
	podLogs := k8s.client.CoreV1().Pods(ns).GetLogs(name, logOptions)

	logStream, err := podLogs.Stream(ctx)

	return logStream, err
}

// ListPodByLabelSelector key1=value1,key2=value2
func (k8s *Kubectl) ListPodByLabelSelector(ctx context.Context, ns, selector string) ([]v1.Pod, error) {
	list, err := k8s.client.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{LabelSelector: selector})
	if err == nil && list != nil && list.Items != nil && len(list.Items) > 0 {
		sort.Slice(list.Items, func(i, j int) bool {
			return list.Items[i].CreationTimestamp.Time.After(list.Items[j].CreationTimestamp.Time)
		})
		return list.Items, nil
	}
	return nil, err
}
