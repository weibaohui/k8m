package node

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

var WebsocketMessageType = map[int]string{
	websocket.BinaryMessage: "binary",
	websocket.TextMessage:   "text",
	websocket.CloseMessage:  "close",
	websocket.PingMessage:   "ping",
	websocket.PongMessage:   "pong",
}

func CreateNodeShell(c *gin.Context) {
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)
	name := c.Param("node_name") // NodeName
	podName, err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Node{}).Name(name).Ctl().Node().CreateNodeShell()

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	ns := "kube-system"

	var p *v1.Pod
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			amis.WriteJsonError(c, fmt.Errorf("等待Pod启动超时"))
			return
		case <-ticker.C:
			err := kom.Cluster(selectedCluster).Resource(&v1.Pod{}).Name(podName).Namespace(ns).Get(&p).Error
			if err != nil {
				klog.V(6).Infof("等待Pod %s/%s 创建中...", ns, podName)
				continue
			}

			if p == nil {
				klog.V(6).Infof("Pod %s/%s 未创建", ns, podName)
				continue
			}

			if len(p.Status.ContainerStatuses) == 0 {
				klog.V(6).Infof("Pod %s/%s 容器状态未就绪", ns, podName)
				continue
			}

			// 检查所有容器是否都Ready
			allContainersReady := true
			for _, status := range p.Status.ContainerStatuses {
				if !status.Ready {
					allContainersReady = false
					klog.V(6).Infof("容器 %s 在Pod %s/%s 中未就绪", status.Name, ns, podName)
					break
				}
			}

			if allContainersReady {
				klog.V(6).Infof("Pod %s/%s 所有容器已就绪", ns, podName)
				break
			}
		}

		// 如果所有容器都Ready，退出循环
		if p != nil && len(p.Status.ContainerStatuses) > 0 {
			allReady := true
			for _, status := range p.Status.ContainerStatuses {
				if !status.Ready {
					allReady = false
					break
				}
			}
			if allReady {
				break
			}
		}

		klog.V(6).Infof("继续等待Pod %s/%s 完全就绪...", ns, podName)
	}

	amis.WriteJsonData(c, gin.H{
		"podName":       podName,
		"ns":            ns,
		"containerName": "shell",
		"pod":           p,
	})
}
