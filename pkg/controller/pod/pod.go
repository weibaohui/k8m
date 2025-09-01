package pod

import (
	"bufio"
	"errors"
	"fmt"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/controller/sse"
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

type LogController struct{}

func RegisterLogRoutes(api *gin.RouterGroup) {
	ctrl := &LogController{}
	api.GET("/pod/logs/sse/ns/:ns/pod_name/:pod_name/container/:container_name", ctrl.StreamLogs)
	api.GET("/pod/logs/download/ns/:ns/pod_name/:pod_name/container/:container_name", ctrl.DownloadLogs)
}

// StreamLogs 通过SSE流式传输Pod日志
// @Summary 流式获取Pod日志
// @Security BearerAuth
// @Param cluster path string true "集群名称"
// @Param ns path string true "命名空间"
// @Param pod_name path string true "Pod名称"
// @Param container_name path string true "容器名称"
// @Success 200 {string} string "日志流"
// @Router /k8s/cluster/{cluster}/pod/logs/sse/ns/{ns}/pod_name/{pod_name}/container/{container_name} [get]
func (lc *LogController) StreamLogs(c *gin.Context) {

	ns := c.Param("ns")
	podName := c.Param("pod_name")
	containerName := c.Param("container_name")

	lop := metav1.ListOptions{}
	labelSelector := c.Request.FormValue("labelSelector")

	//不选pod，设置了labelSelector，说明是要查询多个pod了
	//undefined 是前端处理问题
	if labelSelector != "" && (podName == "" || podName == "undefined") {
		lop.LabelSelector = labelSelector
	}
	// 查某一个pod
	if podName != "" && podName != "undefined" {
		lop.FieldSelector = fmt.Sprintf("metadata.name=%s", podName)
	}
	klog.V(2).Infof("StreamLogs metav1.ListOptions=%v", lop)
	lc.streamPodLogsBySelector(c, ns, containerName, lop)
}
func (lc *LogController) streamPodLogsBySelector(c *gin.Context, ns string, containerName string, options metav1.ListOptions) {
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

	pr, pw := io.Pipe()

	for _, pd := range pods {

		var podName = pd.GetName()
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
		go func(pd *v1.Pod, r io.ReadCloser) {
			defer r.Close()
			var prefix string
			if pd.GetNamespace() != "" {
				prefix = fmt.Sprintf("%s/%s: ", pd.GetNamespace(), pd.GetName())
			}
			if containerName != "" {
				prefix = fmt.Sprintf("%s/%s/%s: ", pd.GetNamespace(), pd.GetName(), containerName)
			}

			scanner := bufio.NewScanner(r)
			for scanner.Scan() {
				line := fmt.Sprintf("[%s] %s", prefix, scanner.Text())
				if _, err := pw.Write([]byte(line)); err != nil {
					return // 管道已关闭
				}
			}
		}(&pd, stream)
	}

	// 监听 ctx，用户断开 SSE 时关闭 PipeWriter
	go func() {
		<-ctx.Done()
		klog.V(2).Infof("SSE connection closed.")
		pw.Close()
	}()

	sse.WriteSSE(c, pr)
}

// DownloadLogs 下载Pod日志
// @Summary 下载Pod日志
// @Security BearerAuth
// @Param cluster path string true "集群名称"
// @Param ns path string true "命名空间"
// @Param pod_name path string true "Pod名称"
// @Param container_name path string true "容器名称"
// @Success 200 {file} file "日志文件"
// @Router /k8s/cluster/{cluster}/pod/logs/download/ns/{ns}/pod_name/{pod_name}/container/{container_name} [get]
func (lc *LogController) DownloadLogs(c *gin.Context) {

	ns := c.Param("ns")
	podName := c.Param("pod_name")
	containerName := c.Param("container_name")
	selector := fmt.Sprintf("metadata.name=%s", podName)
	lc.downloadPodLogsBySelector(c, ns, containerName, metav1.ListOptions{FieldSelector: selector})
}

func (lc *LogController) downloadPodLogsBySelector(c *gin.Context, ns string, containerName string, options metav1.ListOptions) {
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
