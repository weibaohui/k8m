package pod

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/controller/sse"
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func StreamLogs(c *gin.Context) {

	ns := c.Param("ns")
	podName := c.Param("pod_name")
	containerName := c.Param("container_name")
	selector := fmt.Sprintf("metadata.name=%s", podName)
	StreamPodLogsBySelector(c, ns, containerName, metav1.ListOptions{
		FieldSelector: selector,
	})
}
func StreamPodLogsBySelector(c *gin.Context, ns string, containerName string, options metav1.ListOptions) {
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var pods []v1.Pod
	err = kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Pod{}).Namespace(ns).List(&pods, options).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	if len(pods) != 1 {
		amis.WriteJsonError(c, errors.New("pod 数量过多"))
		return
	}

	var podName = pods[0].GetName()
	logOpt, err := BindPodLogOptions(c, containerName)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	podService := service.PodService()
	stream, err := podService.StreamPodLogs(ctx, selectedCluster, ns, podName, logOpt)

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	sse.WriteSSE(c, stream)
}
func DownloadLogs(c *gin.Context) {

	ns := c.Param("ns")
	podName := c.Param("pod_name")
	containerName := c.Param("container_name")
	selector := fmt.Sprintf("metadata.name=%s", podName)
	DownloadPodLogsBySelector(c, ns, containerName, metav1.ListOptions{FieldSelector: selector})
}
 
func DownloadPodLogsBySelector(c *gin.Context, ns string, containerName string, options metav1.ListOptions) {
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	ctx := amis.GetContextWithUser(c)
	var pods []v1.Pod
	err = kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Pod{}).Namespace(ns).List(&pods, options).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	if len(pods) != 1 {
		amis.WriteJsonError(c, errors.New("pod 数量过多"))
		return
	}

	var podName = pods[0].GetName()
	logOpt, err := BindPodLogOptions(c, containerName)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	logOpt.Follow = false

	podService := service.PodService()
	stream, err := podService.StreamPodLogs(ctx, selectedCluster, ns, podName, logOpt)

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	sse.DownloadLog(c, logOpt, stream)
}
func Usage(c *gin.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	usage, err := kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.Pod{}).
		Namespace(ns).
		Name(name).
		Ctl().Pod().ResourceUsageTable()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, usage)
}

// UniqueLabels 返回当前集群中所有唯一的 Pod 标签键列表，格式化为前端可用的选项数组。
func UniqueLabels(c *gin.Context) {
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	labels := service.PodService().GetUniquePodLabels(selectedCluster)

	var names []map[string]string
	for k := range labels {
		names = append(names, map[string]string{
			"label": k,
			"value": k,
		})
	}
	slice.SortBy(names, func(a, b map[string]string) bool {
		return a["label"] < b["label"]
	})
	amis.WriteJsonData(c, gin.H{
		"options": names,
	})
}

// TopList 返回指定命名空间下所有 Pod 的资源使用情况（CPU、内存等），支持多命名空间查询，并以便于前端排序的格式输出。
func TopList(c *gin.Context) {
	ns := c.Param("ns")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	podMetrics, err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Pod{}).
		Namespace(strings.Split(ns, ",")...).
		WithCache(time.Second * 30).
		Ctl().Pod().Top()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 转换为map 前端排序使用，usage.cpu这种前端无法正确排序
	var result []map[string]string
	for _, item := range podMetrics {
		result = append(result, map[string]string{
			"name":            item.Name,
			"namespace":       item.Namespace,
			"cpu":             item.Usage.CPU,
			"memory":          item.Usage.Memory,
			"cpu_nano":        fmt.Sprintf("%d", item.Usage.CPUNano),
			"memory_byte":     fmt.Sprintf("%d", item.Usage.MemoryByte),
			"cpu_fraction":    item.Usage.CPUFraction,
			"memory_fraction": item.Usage.MemoryFraction,
		})
	}
	amis.WriteJsonList(c, result)
}
