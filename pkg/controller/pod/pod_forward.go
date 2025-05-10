package pod

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
)

// PortInfo 结构体用于描述端口转发信息
// 包含容器名、端口名、协议、端口号、本地端口、转发状态等
type PortInfo struct {
	Cluster       string        `json:"cluster"`
	Namespace     string        `json:"namespace"` // Pod 命名空间
	Name          string        `json:"name"`      // pod名称
	ContainerName string        `json:"container_name"`
	PortName      string        `json:"port_name"`  // 端口名称
	Protocol      string        `json:"protocol"`   // TCP/UDP/STCP
	LocalPort     string        `json:"local_port"` // 本地端口，转发端口
	PodPort       string        `json:"pod_port"`   // pod 端口
	Status        string        `json:"status"`     // running/failed/stopped
	StopCh        chan struct{} `json:"-"`
}

// portForwardTable 用于维护所有端口转发的状态
var portForwardTable = make(map[string]*PortInfo) // key: cluster/ns/pod/port
var portForwardTableMutex sync.RWMutex

func StartPortForward(c *gin.Context) {
	ctx := amis.GetContextWithUser(c)
	name := c.Param("name")
	ns := c.Param("ns")
	localPort := c.Param("local_port")
	podPort := c.Param("pod_port")
	containerName := c.Param("container_name")

	// 验证podPort是否为有效的整数
	if _, err := strconv.Atoi(podPort); err != nil {
		amis.WriteJsonError(c, fmt.Errorf("无效的容器组端口号: %s", podPort))
		return
	}

	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 前端是界面点选而来
	// // 检查pod是否存在
	// var pod v1.Pod
	// err = kom.Cluster(selectedCluster).WithContext(ctx).
	// 	Resource(&v1.Pod{}).
	// 	Namespace(ns).
	// 	Name(name).
	// 	Get(&pod).Error
	//
	// if err != nil {
	// 	amis.WriteJsonError(c, err)
	// 	return
	// }

	stopCh := make(chan struct{})
	key := getMapKey(selectedCluster, ns, name, containerName, podPort)

	if localPort == "" {
		localPort = getRandomPort()
	}
	go func() {
		portForwardTableMutex.Lock()
		portForwardTable[key] = &PortInfo{
			Cluster:       selectedCluster,
			Namespace:     ns,
			Name:          name,
			ContainerName: containerName, // 可后续补充
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
			Ctl().Pod().
			ContainerName(containerName).
			PortForward(localPort, podPort, stopCh).Error
		if err != nil {
			portForwardTableMutex.Lock()
			if pf, ok := portForwardTable[key]; ok {
				pf.Status = "failed"
			}
			portForwardTableMutex.Unlock()
		}

	}()
	amis.WriteJsonOK(c)
}
func StopPortForward(c *gin.Context) {
	name := c.Param("name")
	ns := c.Param("ns")
	containerName := c.Param("container_name")
	podPort := c.Param("pod_port")
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	key := getMapKey(selectedCluster, ns, name, containerName, podPort)
	portForwardTableMutex.Lock()

	if pf, ok := portForwardTable[key]; ok {
		pf.StopCh <- struct{}{}
		pf.Status = "stopped"
		pf.LocalPort = ""
	}
	portForwardTableMutex.Unlock()

	amis.WriteJsonOK(c)
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
	var containerPorts []*PortInfo

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
	for _, container := range pod.Spec.Containers {
		for _, port := range container.Ports {
			key := getMapKey(selectedCluster, ns, name, container.Name, fmt.Sprintf("%d", port.ContainerPort))
			status := ""
			localPort := ""
			portForwardTableMutex.RLock()
			if pf, ok := portForwardTable[key]; ok {
				status = pf.Status
				localPort = pf.LocalPort
			}
			portForwardTableMutex.RUnlock()
			portInfo := &PortInfo{
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
	if len(containerPorts) > 0 {
		amis.WriteJsonData(c, containerPorts)
		return
	}

	amis.WriteJsonError(c, fmt.Errorf("无端口数据"))
}

func getMapKey(selectedCluster, ns, name, container, podPort string) string {
	key := fmt.Sprintf("%s/%s/%s/%s/%s", selectedCluster, ns, name, container, podPort)
	return key
}
func getRandomPort() string {
	// 随机取一个端口
	// 如果重复了，就再取一个，直到不重复
	for {
		// TODO 范围 做成一个配置
		port := utils.RandInt(40000, 49999)
		portStr := fmt.Sprintf("%d", port)

		// 检查端口是否已被使用
		portForwardTableMutex.RLock()
		isUsed := false
		for _, portInfo := range portForwardTable {
			if portInfo.LocalPort == portStr {
				isUsed = true
				break
			}
		}
		portForwardTableMutex.RUnlock()

		// 如果端口未被使用，则返回该端口
		if !isUsed {
			return portStr
		}
	}
}
