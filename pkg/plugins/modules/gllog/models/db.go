package models

import (
	"bufio"
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type GlobalLogEntry struct {
	ID          uint      `json:"id"`
	ClusterName string    `json:"cluster_name"`
	Namespace   string    `json:"namespace,omitempty"`
	NodeName    string    `json:"node_name,omitempty"`
	PodName     string    `json:"pod_name,omitempty"`
	Container   string    `json:"container,omitempty"`
	LogLevel    string    `json:"log_level"`
	Source      string    `json:"source"`
	Message     string    `json:"message"`
	Timestamp   time.Time `json:"timestamp"`
}

func parseLogLevel(message string) string {
	patterns := []struct {
		regex *regexp.Regexp
		level string
	}{
		{regexp.MustCompile(`(?i)^\s*\[?(FATAL|CRITICAL)\]?\s*[:：]?`), "FATAL"},
		{regexp.MustCompile(`(?i)^\s*\[?(ERROR|ERR)\]?\s*[:：]?`), "ERROR"},
		{regexp.MustCompile(`(?i)^\s*\[?(WARNING|WARN)\]?\s*[:：]?`), "WARN"},
		{regexp.MustCompile(`(?i)^\s*\[?(INFO|INFORMATION)\]?\s*[:：]?`), "INFO"},
		{regexp.MustCompile(`(?i)^\s*\[?(DEBUG|DBG)\]?\s*[:：]?`), "DEBUG"},
		{regexp.MustCompile(`(?i)^\s*\[?(TRACE|TRC)\]?\s*[:：]?`), "TRACE"},
		{regexp.MustCompile(`(?i)\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2}.*?\s+(FATAL|CRITICAL)\s*[:：]?`), "FATAL"},
		{regexp.MustCompile(`(?i)\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2}.*?\s+(ERROR|ERR)\s*[:：]?`), "ERROR"},
		{regexp.MustCompile(`(?i)\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2}.*?\s+(WARNING|WARN)\s*[:：]?`), "WARN"},
		{regexp.MustCompile(`(?i)\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2}.*?\s+(INFO|INFORMATION)\s*[:：]?`), "INFO"},
		{regexp.MustCompile(`(?i)\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2}.*?\s+(DEBUG|DBG)\s*[:：]?`), "DEBUG"},
		{regexp.MustCompile(`(?i)\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2}:\d{2}.*?\s+(TRACE|TRC)\s*[:：]?`), "TRACE"},
	}

	for _, pattern := range patterns {
		if pattern.regex.MatchString(message) {
			return pattern.level
		}
	}

	if regexp.MustCompile(`(?i)\b(error|failed|exception|panic|crash)\b`).MatchString(message) {
		return "ERROR"
	}
	if regexp.MustCompile(`(?i)\b(warning|warn)\b`).MatchString(message) {
		return "WARN"
	}
	if regexp.MustCompile(`(?i)\b(debug)\b`).MatchString(message) {
		return "DEBUG"
	}

	return "INFO"
}

func ListGlobalLog(ctx context.Context, cluster, namespace, nodeName, podName, container, keyword, logLevel, source, startTime, endTime string) ([]*GlobalLogEntry, error) {
	if cluster == "" {
		return nil, fmt.Errorf("cluster parameter is required")
	}

	if !service.ClusterService().IsConnected(cluster) {
		return nil, fmt.Errorf("cluster %s is not connected", cluster)
	}

	var logs []*GlobalLogEntry
	var err error

	if podName != "" {
		logs, err = queryPodLogs(ctx, cluster, namespace, nodeName, podName, container, keyword, logLevel, source, startTime, endTime)
	} else if namespace != "" || nodeName != "" {
		logs, err = queryFilteredPodsLogs(ctx, cluster, namespace, nodeName, container, keyword, logLevel, source, startTime, endTime)
	} else {
		logs, err = queryClusterLogs(ctx, cluster, container, keyword, logLevel, source, startTime, endTime)
	}

	if err != nil {
		return nil, err
	}

	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Timestamp.After(logs[j].Timestamp)
	})

	return logs, nil
}

func queryPodLogs(ctx context.Context, cluster, namespace, nodeName, podName, container, keyword, logLevel, source, startTime, endTime string) ([]*GlobalLogEntry, error) {
	var pods []v1.Pod
	q := kom.Cluster(cluster).WithContext(ctx).Resource(&v1.Pod{})

	if namespace != "" {
		q = q.Namespace(namespace)
	}

	listOpt := metav1.ListOptions{FieldSelector: fmt.Sprintf("metadata.name=%s", podName)}
	if nodeName != "" {
		listOpt.FieldSelector += fmt.Sprintf(",spec.nodeName=%s", nodeName)
	}

	if err := q.List(&pods, listOpt).Error; err != nil {
		return nil, fmt.Errorf("failed to list pods: %v", err)
	}

	if len(pods) == 0 {
		return []*GlobalLogEntry{}, nil
	}

	var allLogs []*GlobalLogEntry
	tail := int64(200)

	for _, pod := range pods {
		logs, err := fetchPodLogs(ctx, cluster, pod, container, tail)
		if err != nil {
			continue
		}
		allLogs = append(allLogs, filterLogs(logs, keyword, logLevel, source, startTime, endTime)...)
	}

	return allLogs, nil
}

func queryFilteredPodsLogs(ctx context.Context, cluster, namespace, nodeName, container, keyword, logLevel, source, startTime, endTime string) ([]*GlobalLogEntry, error) {
	var pods []v1.Pod
	q := kom.Cluster(cluster).WithContext(ctx).Resource(&v1.Pod{})

	if namespace != "" {
		q = q.Namespace(namespace)
	}

	listOpt := metav1.ListOptions{}
	if nodeName != "" {
		listOpt.FieldSelector = fmt.Sprintf("spec.nodeName=%s", nodeName)
	}

	if err := q.List(&pods, listOpt).Error; err != nil {
		return nil, fmt.Errorf("failed to list pods: %v", err)
	}

	if len(pods) == 0 {
		return []*GlobalLogEntry{}, nil
	}

	if len(pods) > 10 {
		pods = pods[:10]
	}

	var allLogs []*GlobalLogEntry
	tail := int64(100)

	for _, pod := range pods {
		logs, err := fetchPodLogs(ctx, cluster, pod, container, tail)
		if err != nil {
			continue
		}
		allLogs = append(allLogs, filterLogs(logs, keyword, logLevel, source, startTime, endTime)...)
	}

	return allLogs, nil
}

func queryClusterLogs(ctx context.Context, cluster, container, keyword, logLevel, source, startTime, endTime string) ([]*GlobalLogEntry, error) {
	var pods []v1.Pod
	if err := kom.Cluster(cluster).WithContext(ctx).Resource(&v1.Pod{}).List(&pods, metav1.ListOptions{}).Error; err != nil {
		return nil, fmt.Errorf("failed to list pods: %v", err)
	}

	if len(pods) == 0 {
		return []*GlobalLogEntry{}, nil
	}

	if len(pods) > 20 {
		pods = pods[:20]
	}

	var allLogs []*GlobalLogEntry
	tail := int64(50)

	for _, pod := range pods {
		logs, err := fetchPodLogs(ctx, cluster, pod, container, tail)
		if err != nil {
			continue
		}
		allLogs = append(allLogs, filterLogs(logs, keyword, logLevel, source, startTime, endTime)...)
	}

	return allLogs, nil
}

func fetchPodLogs(ctx context.Context, cluster string, pod v1.Pod, container string, tail int64) ([]*GlobalLogEntry, error) {
	opt := &v1.PodLogOptions{TailLines: &tail}
	if container != "" {
		opt.Container = container
	}

	stream, err := service.PodService().StreamPodLogs(ctx, cluster, pod.Namespace, pod.Name, opt)
	if err != nil || stream == nil {
		return nil, err
	}
	defer stream.Close()

	var logs []*GlobalLogEntry
	scanner := bufio.NewScanner(stream)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 10*1024*1024)

	id := uint(1)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		logs = append(logs, &GlobalLogEntry{
			ID:          id,
			ClusterName: cluster,
			Namespace:   pod.Namespace,
			NodeName:    pod.Spec.NodeName,
			PodName:     pod.Name,
			Container:   container,
			LogLevel:    parseLogLevel(line),
			Source:      "pod",
			Message:     line,
			Timestamp:   time.Now(),
		})
		id++
	}

	return logs, scanner.Err()
}

func filterLogs(logs []*GlobalLogEntry, keyword, logLevel, source, startTime, endTime string) []*GlobalLogEntry {
	var filtered []*GlobalLogEntry

	var startT, endT time.Time
	var err error
	if startTime != "" {
		startT, err = time.Parse("2006-01-02 15:04:05", startTime)
		if err != nil {
			startT = time.Time{}
		}
	}
	if endTime != "" {
		endT, err = time.Parse("2006-01-02 15:04:05", endTime)
		if err != nil {
			endT = time.Time{}
		}
	}

	for _, log := range logs {
		if keyword != "" && !strings.Contains(log.Message, keyword) {
			continue
		}

		if logLevel != "" && log.LogLevel != logLevel {
			continue
		}

		if source != "" && log.Source != source {
			continue
		}

		if !startT.IsZero() && log.Timestamp.Before(startT) {
			continue
		}
		if !endT.IsZero() && log.Timestamp.After(endT) {
			continue
		}

		filtered = append(filtered, log)
	}

	return filtered
}
