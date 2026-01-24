package svc

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/controller/pod"
	"github.com/weibaohui/k8m/pkg/response"
	"github.com/weibaohui/kom/kom"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/klog/v2"
)

type PortForwardController struct{}

type svcPortForwardEntry struct {
	Cluster       string
	Namespace     string
	ServiceName   string
	ServicePort   string
	PodName       string
	ContainerName string
	PodPort       string
}

type ServicePortForwardItem struct {
	PortName    string `json:"port_name"`
	ServicePort string `json:"svc_port"`
	TargetPort  string `json:"target_port"`
	Protocol    string `json:"protocol"`
	LocalPort   string `json:"local_port"`
	Status      string `json:"status"`
	PodName     string `json:"pod_name"`
}

var svcPortForwardTable = map[string]*svcPortForwardEntry{}
var svcPortForwardTableMutex sync.RWMutex

// RegisterPortForwardRoutes 注册 Service 端口转发相关路由。
func RegisterPortForwardRoutes(r chi.Router) {
	ctrl := &PortForwardController{}
	r.Post("/service/port_forward/ns/{ns}/name/{name}/svc_port/{svc_port}/start", response.Adapter(ctrl.StartPortForward))
	r.Post("/service/port_forward/ns/{ns}/name/{name}/svc_port/{svc_port}/stop", response.Adapter(ctrl.StopPortForward))
	r.Get("/service/port_forward/ns/{ns}/name/{name}/port/list", response.Adapter(ctrl.PortForwardList))
}

// StartPortForward 根据 Service 的 selector 找到第一个 Pod，并发起端口转发。
func (pc *PortForwardController) StartPortForward(c *response.Context) {
	ctx := amis.GetContextWithUser(c)
	ns := c.Param("ns")
	svcName := c.Param("name")
	svcPortStr := c.Param("svc_port")

	_, err := strconv.Atoi(svcPortStr)
	if err != nil {
		amis.WriteJsonError(c, fmt.Errorf("无效的 Service 端口号: %s", svcPortStr))
		return
	}

	var req struct {
		LocalPort string `json:"local_port"`
	}
	_ = c.ShouldBindJSON(&req)
	localPort := req.LocalPort
	if localPort == "undefined" {
		localPort = ""
	}

	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var svc corev1.Service
	err = kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&corev1.Service{}).
		Namespace(ns).
		Name(svcName).
		Get(&svc).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	if len(svc.Spec.Selector) == 0 {
		amis.WriteJsonError(c, fmt.Errorf("Service %s/%s 未配置 selector，无法进行端口转发", ns, svcName))
		return
	}

	labelSelector := labels.SelectorFromSet(svc.Spec.Selector).String()
	var pods []corev1.Pod
	err = kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&corev1.Pod{}).
		Namespace(ns).
		List(&pods, metav1.ListOptions{LabelSelector: labelSelector}).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	if len(pods) == 0 {
		amis.WriteJsonError(c, fmt.Errorf("未找到匹配 Service selector 的 Pod，selector=%s", labelSelector))
		return
	}
	targetPod := pods[0]

	svcPort, targetPort, protocol, err := getServicePortInfo(&svc, svcPortStr)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	containerName, podPort, err := resolvePodPortAndContainer(&targetPod, targetPort, svcPort)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	klog.V(6).Infof("Service端口转发开始 集群=%s Service=%s/%s selector=%s Pod=%s 容器=%s 端口=%s 协议=%s", selectedCluster, ns, svcName, labelSelector, targetPod.Name, containerName, podPort, protocol)

	tableKey := getSvcForwardMapKey(selectedCluster, ns, svcName, svcPortStr)
	svcPortForwardTableMutex.Lock()
	if old, ok := svcPortForwardTable[tableKey]; ok {
		pod.StopPortForwardByPod(selectedCluster, ns, old.PodName, old.ContainerName, old.PodPort)
	}
	svcPortForwardTable[tableKey] = &svcPortForwardEntry{
		Cluster:       selectedCluster,
		Namespace:     ns,
		ServiceName:   svcName,
		ServicePort:   svcPortStr,
		PodName:       targetPod.Name,
		ContainerName: containerName,
		PodPort:       podPort,
	}
	svcPortForwardTableMutex.Unlock()

	_, err = pod.StartPortForwardByPod(ctx, selectedCluster, ns, targetPod.Name, containerName, podPort, localPort)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	amis.WriteJsonOK(c)
}

// StopPortForward 停止 Service 对应端口的转发。
func (pc *PortForwardController) StopPortForward(c *response.Context) {
	ns := c.Param("ns")
	svcName := c.Param("name")
	svcPortStr := c.Param("svc_port")

	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	tableKey := getSvcForwardMapKey(selectedCluster, ns, svcName, svcPortStr)

	svcPortForwardTableMutex.RLock()
	entry, ok := svcPortForwardTable[tableKey]
	svcPortForwardTableMutex.RUnlock()
	if ok {
		klog.V(6).Infof("Service端口转发停止 集群=%s Service=%s/%s svcPort=%s Pod=%s 容器=%s 端口=%s", selectedCluster, ns, svcName, svcPortStr, entry.PodName, entry.ContainerName, entry.PodPort)
		pod.StopPortForwardByPod(selectedCluster, ns, entry.PodName, entry.ContainerName, entry.PodPort)
	}

	amis.WriteJsonOK(c)
}

// PortForwardList 返回 Service 各端口的转发状态列表。
func (pc *PortForwardController) PortForwardList(c *response.Context) {
	ctx := amis.GetContextWithUser(c)
	ns := c.Param("ns")
	svcName := c.Param("name")

	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var svc corev1.Service
	err = kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&corev1.Service{}).
		Namespace(ns).
		Name(svcName).
		Get(&svc).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var items []*ServicePortForwardItem
	for _, sp := range svc.Spec.Ports {
		svcPortStr := fmt.Sprintf("%d", sp.Port)
		targetPortStr := sp.TargetPort.String()
		if targetPortStr == "" || targetPortStr == "0" {
			targetPortStr = svcPortStr
		}
		item := &ServicePortForwardItem{
			PortName:    sp.Name,
			ServicePort: svcPortStr,
			TargetPort:  targetPortStr,
			Protocol:    string(sp.Protocol),
		}

		tableKey := getSvcForwardMapKey(selectedCluster, ns, svcName, svcPortStr)
		svcPortForwardTableMutex.RLock()
		entry, ok := svcPortForwardTable[tableKey]
		svcPortForwardTableMutex.RUnlock()
		if ok {
			status, localPort, _ := pod.GetPortForwardStatus(selectedCluster, ns, entry.PodName, entry.ContainerName, entry.PodPort)
			item.Status = status
			item.LocalPort = localPort
			item.PodName = entry.PodName
		}

		items = append(items, item)
	}

	amis.WriteJsonData(c, items)
}

// getSvcForwardMapKey 用于生成 Service 端口转发的映射 Key。
func getSvcForwardMapKey(cluster, ns, svcName, svcPort string) string {
	return fmt.Sprintf("%s/%s/%s/%s", cluster, ns, svcName, svcPort)
}

// getServicePortInfo 解析 Service 端口配置，返回匹配的端口与 targetPort 信息。
func getServicePortInfo(svc *corev1.Service, svcPort string) (servicePort string, targetPort intstr.IntOrString, protocol string, err error) {
	svcPortInt, err := strconv.Atoi(svcPort)
	if err != nil {
		return "", intstr.IntOrString{}, "", fmt.Errorf("无效的 Service 端口号: %s", svcPort)
	}
	for _, p := range svc.Spec.Ports {
		if int(p.Port) == svcPortInt {
			return svcPort, p.TargetPort, string(p.Protocol), nil
		}
	}
	return "", intstr.IntOrString{}, "", fmt.Errorf("Service 未找到端口配置: %s", svcPort)
}

// resolvePodPortAndContainer 将 Service 的 targetPort 映射为 Pod 的容器与容器端口。
// 当 targetPort 为端口名时，优先按端口名匹配；当为端口号时按 ContainerPort 匹配；
// 若 Pod 未声明 Ports，则回退选择第一个容器并使用端口号。
func resolvePodPortAndContainer(p *corev1.Pod, targetPort intstr.IntOrString, defaultPort string) (containerName, podPort string, err error) {
	if p == nil {
		return "", "", fmt.Errorf("Pod 为空")
	}

	if len(p.Spec.Containers) == 0 {
		return "", "", fmt.Errorf("Pod %s 无容器信息", p.Name)
	}

	if targetPort.Type == intstr.String && targetPort.StrVal != "" {
		for _, c := range p.Spec.Containers {
			for _, cp := range c.Ports {
				if cp.Name == targetPort.StrVal {
					return c.Name, fmt.Sprintf("%d", cp.ContainerPort), nil
				}
			}
		}
		return "", "", fmt.Errorf("Pod %s 未找到名为 %s 的容器端口", p.Name, targetPort.StrVal)
	}

	portInt := targetPort.IntValue()
	if portInt == 0 {
		portInt, _ = strconv.Atoi(defaultPort)
	}
	if portInt == 0 {
		return "", "", fmt.Errorf("目标端口为空，无法转发")
	}

	for _, c := range p.Spec.Containers {
		for _, cp := range c.Ports {
			if int(cp.ContainerPort) == portInt {
				return c.Name, fmt.Sprintf("%d", cp.ContainerPort), nil
			}
		}
	}

	return p.Spec.Containers[0].Name, fmt.Sprintf("%d", portInt), nil
}
