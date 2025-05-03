package pod

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/controller/sse"
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
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

func WsExec(c *gin.Context) {

	ns := c.Param("ns")
	podName := c.Param("pod_name")
	containerName := c.Param("container_name")
	cmd := c.Query("cmd")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	if cmd == "" {
		amis.WriteJsonError(c, fmt.Errorf("执行命令为空"))
		return
	}

	// 定义 WebSocket 升级器
	var upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			// 允许所有来源
			return true
		},
	}

	// 将 HTTP 连接升级为 WebSocket 连接
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		klog.Errorf("WebSocket Upgrade Error:%v", err)
		return
	}
	defer conn.Close()
	klog.V(6).Infof("ws Client connected")

	klog.V(6).Infof("cmd=%s\n", cmd)

	cb := func(data []byte) error {
		// 发送数据给客户端
		conn.WriteJSON(gin.H{
			"data": string(data),
		})
		return nil
	}
	err = kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.Pod{}).
		Namespace(ns).
		Name(podName).Ctl().Pod().
		ContainerName(containerName).
		Command("sh", "-c", cmd).
		StreamExecute(cb, cb).Error

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	select {}
}
func Exec(c *gin.Context) {
	ns := c.Param("ns")
	podName := c.Param("pod_name")
	containerName := c.Param("container_name")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	ctx, cancel := context.WithDeadline(ctx, time.Now().Add(time.Second*5))
	defer cancel()
	// 初始化结构体实例
	var payload struct {
		Command string `json:"cmd"`
	}

	// 反序列化 JSON 数据到结构体
	if err := c.ShouldBindJSON(&payload); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	var humanCommand string
	chatService := service.ChatService()
	humanCommand = payload.Command
	if service.AIService().IsEnabled() {
		prompt := fmt.Sprintf("请根据用户描述，给出最合适的一条命令。第一步，给出命令，第二步，检查命令是否为单行单个命令。请务必注意，只给出一条命令。请不要使用top、tail -f等流式输出的命令，请要不使用tzdate等交互性的命令。只能使用输入命令，紧接着输出完整返回的命令。请不要做任何解释。最终的代码一定、务必、必须用```bash\n命令写这里\n```包裹起来\n以下为用户的要求:\n%s", strings.TrimPrefix(payload.Command, "#"))
		aiCmd := chatService.Chat(prompt)
		cleanCmd := chatService.CleanCmd(aiCmd)
		klog.V(4).Infof("\n用户输入:\t%s\nprompt:\t%s\nAI返回:\t%s\n提取命令:\t%s\n", payload.Command, prompt, aiCmd, cleanCmd)
		if cleanCmd != "" {
			payload.Command = cleanCmd
		}
	} else {
		// 未开启，那么删除掉#,提高容错处理
		humanCommand = strings.TrimPrefix(payload.Command, "#")
		payload.Command = humanCommand
	}

	var result []byte
	err = kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.Pod{}).
		Namespace(ns).
		Name(podName).Ctl().Pod().
		ContainerName(containerName).Command("sh", "-c", payload.Command).Execute(&result).Error

	if err != nil {
		amis.WriteJsonData(c, gin.H{
			"result":        fmt.Sprintf("%v", err.Error()),
			"human_command": humanCommand,
			"last_command":  payload.Command,
		})
		return
	}
	amis.WriteJsonData(c, gin.H{
		"result":        string(result),
		"human_command": humanCommand,
		"last_command":  payload.Command,
	})

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
