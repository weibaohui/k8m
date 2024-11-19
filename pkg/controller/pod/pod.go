package pod

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
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
	ctx := c.Request.Context()

	var pods []v1.Pod
	err := kom.DefaultCluster().Resource(&v1.Pod{}).Namespace(ns).List(&pods, options).Error
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
	podService := &service.PodService{}
	stream, err := podService.StreamPodLogs(ctx, ns, podName, logOpt)

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
func Exec(c *gin.Context) {
	ns := c.Param("ns")
	podName := c.Param("pod_name")
	containerName := c.Param("container_name")
	ctx := c.Request.Context()

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
	chatService := service.ChatServiceInstance()
	humanCommand = payload.Command
	if chatService.IsEnabled() {
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
	err := kom.DefaultCluster().WithContext(ctx).
		Resource(&v1.Pod{}).
		Namespace(ns).
		Name(podName).
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
	ctx := c.Request.Context()
	var pods []v1.Pod
	err := kom.DefaultCluster().Resource(&v1.Pod{}).Namespace(ns).List(&pods, options).Error
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

	podService := &service.PodService{}
	stream, err := podService.StreamPodLogs(ctx, ns, podName, logOpt)

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	sse.DownloadLog(c, logOpt, stream)
}
