package pod

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/controller/sse"
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"k8s.io/utils/strings/slices"
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
	selectedCluster := amis.GetSelectedCluster(c)

	var pods []v1.Pod
	err := kom.Cluster(selectedCluster).Resource(&v1.Pod{}).Namespace(ns).List(&pods, options).Error
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
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

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
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

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
	err := kom.Cluster(selectedCluster).WithContext(ctx).
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
	selectedCluster := amis.GetSelectedCluster(c)

	ctx := c.Request.Context()
	var pods []v1.Pod
	err := kom.Cluster(selectedCluster).Resource(&v1.Pod{}).Namespace(ns).List(&pods, options).Error
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
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	usage := kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.Pod{}).
		Namespace(ns).
		Name(name).
		Ctl().Pod().ResourceUsageTable()
	amis.WriteJsonData(c, usage)
}
func LinksServices(c *gin.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	var pod v1.Pod
	err := kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.Pod{}).
		Namespace(ns).
		Name(name).
		WithCache(24 * time.Hour).
		Get(&pod).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	services, err := getServicesByPodLabels(ctx, selectedCluster, &pod)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonList(c, services)
}

// getServicesByPodLabels 获取与Pod关联的Service
func getServicesByPodLabels(ctx context.Context, selectedCluster string, pod *v1.Pod) ([]v1.Service, error) {
	// 	查询流程
	// 获取目标 Pod 的详细信息：

	// 使用 Pod 的 API 获取其 metadata.labels。
	// 确定 Pod 所在的 Namespace。
	// 获取 Namespace 内的所有 Services：

	// 使用 kubectl get services -n {namespace} 或调用 API /api/v1/namespaces/{namespace}/services。
	// 逐个匹配 Service 的 selector：

	// 对每个 Service：
	// 提取其 spec.selector。
	// 遍历 selector 的所有键值对，检查 Pod 是否包含这些标签且值相等。
	// 如果所有标签条件都满足，将此 Service 记录为与该 Pod 关联。
	// 返回结果：

	// 将所有匹配的 Service 名称及相关信息返回。

	podLabels := pod.GetLabels()

	var services []v1.Service
	err := kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.Service{}).
		Namespace(pod.Namespace).
		WithCache(3 * time.Minute). // 3分钟缓存，不宜过久
		List(&services).Error
	if err != nil {
		return nil, err
	}

	var result []v1.Service
	for _, svc := range services {
		serviceLabels := svc.Spec.Selector
		// 遍历selector
		// serviceLabels中所有的kv,都必须在podLabels中存在,且值相等
		// 如果有一个不满足,则跳过
		for k, v := range serviceLabels {
			if podLabels[k] != v {
				continue
			}
			result = append(result, svc)
		}

	}
	return result, nil
}
func LinksEndpoints(c *gin.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	var pod v1.Pod
	err := kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.Pod{}).
		Namespace(ns).
		Name(name).
		WithCache(24 * time.Hour).
		Get(&pod).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	services, err := getServicesByPodLabels(ctx, selectedCluster, &pod)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// endpoints 与 svc 同名
	// 1.获取service 名称
	// 2.获取endpoints
	// 3.返回endpoints

	var names []string
	for _, svc := range services {
		names = append(names, svc.Name)
	}

	var endpoints []v1.Endpoints

	for _, name := range names {
		var endpoint v1.Endpoints
		err = kom.Cluster(selectedCluster).WithContext(ctx).
			Resource(&v1.Endpoints{}).
			Namespace(ns).
			Name(name).
			Get(&endpoint).Error
		if err != nil {
			continue
		}
		endpoints = append(endpoints, endpoint)
	}

	amis.WriteJsonList(c, endpoints)

}

func LinksPVC(c *gin.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	var pod v1.Pod
	err := kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.Pod{}).
		Namespace(ns).
		Name(name).
		WithCache(24 * time.Hour).
		Get(&pod).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	// 找打pvc 名称列表
	var pvcNames []string
	for _, volume := range pod.Spec.Volumes {
		if volume.PersistentVolumeClaim != nil {
			pvcNames = append(pvcNames, volume.PersistentVolumeClaim.ClaimName)
		}
	}

	// 找出同ns下pvc的列表，过滤pvcNames
	var pvcList []v1.PersistentVolumeClaim
	err = kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.PersistentVolumeClaim{}).
		Namespace(ns).
		WithCache(24 * time.Hour).
		List(&pvcList).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 过滤pvcList，只保留pvcNames
	var result []v1.PersistentVolumeClaim
	for _, pvc := range pvcList {
		if slices.Contains(pvcNames, pvc.Name) {
			result = append(result, pvc)
		}
	}

	amis.WriteJsonList(c, result)
}

func LinksIngress(c *gin.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	var pod v1.Pod
	err := kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.Pod{}).
		Namespace(ns).
		Name(name).
		WithCache(24 * time.Hour).
		Get(&pod).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	services, err := getServicesByPodLabels(ctx, selectedCluster, &pod)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var servicesName []string
	for _, svc := range services {
		servicesName = append(servicesName, svc.Name)
	}

	// 获取ingress
	// Ingress 通过 spec.rules 或 spec.defaultBackend 中的 service.name 指定关联的 Service。
	// 遍历services，获取ingress
	var ingressList []networkingv1.Ingress
	err = kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&networkingv1.Ingress{}).
		Namespace(ns).
		WithCache(24 * time.Hour).
		List(&ingressList).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 过滤ingressList，只保留与services关联的ingress
	var result []networkingv1.Ingress
	for _, ingress := range ingressList {
		if slices.Contains(servicesName, ingress.Spec.Rules[0].Host) {
			result = append(result, ingress)
		}
	}
	// 遍历 Ingress 检查关联
	for _, ingress := range ingressList {
		if ingress.Spec.DefaultBackend != nil {
			if ingress.Spec.DefaultBackend.Service != nil && ingress.Spec.DefaultBackend.Service.Name != "" {
				if slices.Contains(servicesName, ingress.Spec.DefaultBackend.Service.Name) {
					result = append(result, ingress)
				}
			}
		}

		for _, rule := range ingress.Spec.Rules {
			if rule.HTTP != nil {
				for _, path := range rule.HTTP.Paths {
					if path.Backend.Service != nil && path.Backend.Service.Name != "" {

						backName := path.Backend.Service.Name
						if slices.Contains(servicesName, backName) {
							result = append(result, ingress)
						}
					}
				}

			}

		}
	}

	amis.WriteJsonList(c, result)

}
