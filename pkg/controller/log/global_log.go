package log

import (
	"bufio"
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GlobalLogEntry 全局日志条目
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

// parseLogLevel 从日志消息中解析日志级别
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

	// 优先检查明确的日志级别标识
	for _, pattern := range patterns {
		if pattern.regex.MatchString(message) {
			return pattern.level
		}
	}

	// 检查是否包含错误相关的独立词汇
	if regexp.MustCompile(`(?i)\b(error|failed|exception|panic|crash)\b`).MatchString(message) {
		return "ERROR"
	}
	if regexp.MustCompile(`(?i)\b(warning|warn)\b`).MatchString(message) {
		return "WARN"
	}
	if regexp.MustCompile(`(?i)\b(debug)\b`).MatchString(message) {
		return "DEBUG"
	}

	// 默认返回 INFO
	return "INFO"
}

// ListGlobalLog 全局日志列表
func (lc *Controller) ListGlobalLog(c *gin.Context) {
	cluster := c.Query("cluster")
	namespace := c.Query("namespace")
	nodeName := c.Query("node_name")
	podName := c.Query("pod_name")
	container := c.Query("container")
	keyword := c.Query("keyword")
	logLevel := c.Query("log_level")
	source := c.Query("source")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	// 集群是必需的
	if cluster == "" {
		amis.WriteJsonError(c, fmt.Errorf("cluster parameter is required"))
		return
	}

	// 手动设置集群到上下文中，因为 /mgm/ 路径不会经过集群中间件
	c.Set("cluster", cluster)

	// 检查集群是否连接
	if !service.ClusterService().IsConnected(cluster) {
		amis.WriteJsonError(c, fmt.Errorf("cluster %s is not connected", cluster))
		return
	}

	ctx := amis.GetContextWithUser(c)

	// 使用 GetSelectedCluster 来获取集群，保持与其他 API 的一致性
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	var logs []*GlobalLogEntry

	// 根据查询参数决定查询策略
	if podName != "" {
		// 查询特定 Pod 的日志
		logs, err = lc.queryPodLogs(ctx, selectedCluster, namespace, nodeName, podName, container, keyword, logLevel, source, startTime, endTime)
	} else if namespace != "" || nodeName != "" {
		// 查询特定命名空间或节点下的 Pod 日志
		logs, err = lc.queryFilteredPodsLogs(ctx, selectedCluster, namespace, nodeName, container, keyword, logLevel, source, startTime, endTime)
	} else {
		// 查询整个集群的 Pod 日志（限制数量）
		logs, err = lc.queryClusterLogs(ctx, selectedCluster, container, keyword, logLevel, source, startTime, endTime)
	}

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 按时间倒序排序
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Timestamp.After(logs[j].Timestamp)
	})

	amis.WriteJsonListWithTotal(c, int64(len(logs)), logs)
}

// queryPodLogs 查询特定 Pod 的日志
func (lc *Controller) queryPodLogs(ctx context.Context, cluster, namespace, nodeName, podName, container, keyword, logLevel, source, startTime, endTime string) ([]*GlobalLogEntry, error) {
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
	tail := int64(200) // 单个 Pod 获取更多日志

	for _, pod := range pods {
		logs, err := lc.fetchPodLogs(ctx, cluster, pod, container, tail)
		if err != nil {
			continue // 跳过错误的 Pod，继续处理其他
		}
		allLogs = append(allLogs, lc.filterLogs(logs, keyword, logLevel, source, startTime, endTime)...)
	}

	return allLogs, nil
}

// queryFilteredPodsLogs 查询特定命名空间或节点下的 Pod 日志
func (lc *Controller) queryFilteredPodsLogs(ctx context.Context, cluster, namespace, nodeName, container, keyword, logLevel, source, startTime, endTime string) ([]*GlobalLogEntry, error) {
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

	// 限制 Pod 数量以避免过多请求
	if len(pods) > 10 {
		pods = pods[:10]
	}

	var allLogs []*GlobalLogEntry
	tail := int64(100)

	for _, pod := range pods {
		logs, err := lc.fetchPodLogs(ctx, cluster, pod, container, tail)
		if err != nil {
			continue
		}
		allLogs = append(allLogs, lc.filterLogs(logs, keyword, logLevel, source, startTime, endTime)...)
	}

	return allLogs, nil
}

// queryClusterLogs 查询整个集群的 Pod 日志
func (lc *Controller) queryClusterLogs(ctx context.Context, cluster, container, keyword, logLevel, source, startTime, endTime string) ([]*GlobalLogEntry, error) {
	var pods []v1.Pod
	if err := kom.Cluster(cluster).WithContext(ctx).Resource(&v1.Pod{}).List(&pods, metav1.ListOptions{}).Error; err != nil {
		return nil, fmt.Errorf("failed to list pods: %v", err)
	}

	if len(pods) == 0 {
		return []*GlobalLogEntry{}, nil
	}

	// 严格限制全集群查询的 Pod 数量
	if len(pods) > 20 {
		pods = pods[:20]
	}

	var allLogs []*GlobalLogEntry
	tail := int64(50)

	for _, pod := range pods {
		logs, err := lc.fetchPodLogs(ctx, cluster, pod, container, tail)
		if err != nil {
			continue
		}
		allLogs = append(allLogs, lc.filterLogs(logs, keyword, logLevel, source, startTime, endTime)...)
	}

	return allLogs, nil
}

// fetchPodLogs 从单个 Pod 获取日志
func (lc *Controller) fetchPodLogs(ctx context.Context, cluster string, pod v1.Pod, container string, tail int64) ([]*GlobalLogEntry, error) {
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

// filterLogs 根据条件过滤日志
func (lc *Controller) filterLogs(logs []*GlobalLogEntry, keyword, logLevel, source, startTime, endTime string) []*GlobalLogEntry {
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
		// 关键字过滤
		if keyword != "" && !strings.Contains(log.Message, keyword) {
			continue
		}

		// 日志级别过滤
		if logLevel != "" && log.LogLevel != logLevel {
			continue
		}

		// 来源过滤
		if source != "" && log.Source != source {
			continue
		}

		// 时间范围过滤
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
