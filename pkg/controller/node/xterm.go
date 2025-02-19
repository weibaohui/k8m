package node

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/core/v1"
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
	ns, podName, containerName, err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Node{}).Name(name).Ctl().Node().CreateNodeShell()

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var p *v1.Pod
	err = kom.Cluster(selectedCluster).Resource(&v1.Pod{}).Name(podName).Namespace(ns).Get(&p).Error

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, gin.H{
		"podName":       podName,
		"ns":            ns,
		"containerName": containerName,
		"pod":           p,
	})
}

func CreateKubectlShell(c *gin.Context) {
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)
	name := c.Param("node_name") // NodeName

	kubeconfig := service.ClusterService().GetClusterByID("kubeconfig.yaml/8o2u742o@sealos").GetKubeconfig()
	ns, podName, containerName, err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Node{}).Name(name).Ctl().Node().CreateKubectlShell(kubeconfig)

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var p *v1.Pod
	err = kom.Cluster(selectedCluster).Resource(&v1.Pod{}).Name(podName).Namespace(ns).Get(&p).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, gin.H{
		"podName":       podName,
		"ns":            ns,
		"containerName": containerName,
		"pod":           p,
	})
}
