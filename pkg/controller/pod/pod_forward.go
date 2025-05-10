package pod

import (
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
)

// PortInfo 结构体用于描述端口转发信息
// 包含容器名、端口名、协议、端口号、本地端口、转发状态等
type PortInfo struct {
	ContainerName string        `json:"container_name"`
	PortName      string        `json:"port_name"`
	Protocol      string        `json:"protocol"`
	LocalPort     string        `json:"local_port"`
	PodPort       string        `json:"pod_port"`
	Status        string        `json:"status"` // running/failed/stopped
	StopCh        chan struct{} `json:"-"`
}

// portForwardTable 用于维护所有端口转发的状态
var portForwardTable = make(map[string]*PortInfo) // key: cluster/ns/pod/port
var portForwardTableMutex sync.RWMutex

func PortForward(c *gin.Context) {
	ctx := amis.GetContextWithUser(c)
	name := c.Param("name")
	ns := c.Param("ns")
	localPort := c.Param("localPort")
	podPort := c.Param("podPort")
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	stopCh := make(chan struct{})
	key := fmt.Sprintf("%s/%s/%s/%s", selectedCluster, ns, name, podPort)
	portForwardTableMutex.Lock()
	portForwardTable[key] = &PortInfo{
		ContainerName: "", // 可后续补充
		PortName:      "",
		Protocol:      "",
		LocalPort:     localPort,
		PodPort:       podPort,
		Status:        "running",
		StopCh:        stopCh,
	}
	portForwardTableMutex.Unlock()
	err = kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.Pod{}).
		Namespace(ns).
		Name(name).
		PortForward(localPort, podPort, stopCh).Error
	if err != nil {
		portForwardTableMutex.Lock()
		if pf, ok := portForwardTable[key]; ok {
			pf.Status = "failed"
		}
		portForwardTableMutex.Unlock()
		amis.WriteJsonError(c, err)
		return
	}
	// 正常结束后可设置为 stopped
	portForwardTableMutex.Lock()
	if pf, ok := portForwardTable[key]; ok {
		pf.Status = "stopped"
	}
	portForwardTableMutex.Unlock()
}

func PortForwardList(c *gin.Context) {
	ctx := amis.GetContextWithUser(c)
	name := c.Param("name")
	ns := c.Param("ns")
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	var pod *v1.Pod
	err = kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.Pod{}).
		Namespace(ns).
		Name(name).
		Get(&pod).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	var containerPorts []PortInfo
	for _, container := range pod.Spec.Containers {
		for _, port := range container.Ports {
			key := fmt.Sprintf("%s/%s/%s/%d", selectedCluster, ns, name, port.ContainerPort)
			status := ""
			localPort := ""
			portForwardTableMutex.RLock()
			if pf, ok := portForwardTable[key]; ok {
				status = pf.Status
				localPort = pf.LocalPort
			}
			portForwardTableMutex.RUnlock()
			portInfo := PortInfo{
				ContainerName: container.Name,
				PortName:      port.Name,
				Protocol:      string(port.Protocol),
				LocalPort:     localPort,
				PodPort:       fmt.Sprintf("%d", port.ContainerPort),
				Status:        status,
			}
			containerPorts = append(containerPorts, portInfo)
		}
	}
	amis.WriteJsonData(c, containerPorts)
}
