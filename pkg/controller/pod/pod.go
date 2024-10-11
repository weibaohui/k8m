package pod

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/internal/kubectl"
	"github.com/weibaohui/k8m/internal/utils/amis"
	"github.com/weibaohui/k8m/pkg/controller/sse"
)

func StreamLogs(c *gin.Context) {

	var ns = c.Param("ns")
	var podName = c.Param("pod_name")
	var containerName = c.Param("container_name")
	selector := fmt.Sprintf("metadata.name=%s", podName)
	StreamPodLogsBySelector(c, ns, containerName, kubectl.WithFieldSelector(selector))
}
func StreamPodLogsBySelector(c *gin.Context, ns string, containerName string, opts ...kubectl.ListOption) {
	pods, err := kubectl.Init().ListResources(kubectl.Pod, ns, opts...)
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
	stream, err := kubectl.Init().StreamPodLogs(ns, podName, logOpt)

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	sse.WriteSSE(c, stream)
}
func DownloadLogs(c *gin.Context) {

	var ns = c.Param("ns")
	var podName = c.Param("pod_name")
	var containerName = c.Param("container_name")
	selector := fmt.Sprintf("metadata.name=%s", podName)
	DownloadPodLogsBySelector(c, ns, containerName, kubectl.WithFieldSelector(selector))
}
func DownloadPodLogsBySelector(c *gin.Context, ns string, containerName string, opts ...kubectl.ListOption) {
	pods, err := kubectl.Init().ListResources(kubectl.Pod, ns, opts...)
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

	stream, err := kubectl.Init().StreamPodLogs(ns, podName, logOpt)

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	sse.DownloadLog(c, logOpt, stream)
}
