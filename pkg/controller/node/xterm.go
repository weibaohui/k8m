package node

import (
	"encoding/base64"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/flag"
	"github.com/weibaohui/k8m/pkg/service"
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
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	name := c.Param("node_name") // NodeName
	cfg := flag.Init()
	timeout := cfg.ImagePullTimeout
	klog.V(6).Infof("CreateNodeShell timeout: %v", timeout)
	ns, podName, containerName, err := kom.Cluster(selectedCluster).WithContext(ctx).WithCache(time.Duration(timeout) * time.Second).Resource(&v1.Node{}).Name(name).Ctl().Node().CreateNodeShell(cfg.NodeShellImage)

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var p *v1.Pod
	err = kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Pod{}).Name(podName).Namespace(ns).Get(&p).Error

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
	ctx := amis.GetContextWithUser(c)
	name := c.Param("node_name")             // NodeName
	clusterIDBase64 := c.Param("cluster_id") // 集群ID，base64编码
	// base64 解码
	clusterIDBytes, err := base64.StdEncoding.DecodeString(clusterIDBase64)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	clusterID := string(clusterIDBytes)

	if clusterID == "" {
		amis.WriteJsonError(c, fmt.Errorf("集群ID不能为空"))
		return
	}

	// 当前限制为kubectl 安装到本集群中，那么要先检查是否可连接。
	if !service.ClusterService().IsConnected(clusterID) {
		amis.WriteJsonError(c, fmt.Errorf("集群%s 不可用,请先连接该集群，然后重试", clusterID))
	}

	kubeconfig := service.ClusterService().GetClusterByID(string(clusterID)).GetKubeconfig()
	cfg := flag.Init()
	timeout := cfg.ImagePullTimeout
	klog.V(6).Infof("CreateKubectlShell timeout: %v", timeout)
	ns, podName, containerName, err := kom.Cluster(clusterID).WithContext(ctx).WithCache(time.Duration(timeout)*time.Second).Resource(&v1.Node{}).Name(name).Ctl().Node().CreateKubectlShell(kubeconfig, cfg.KubectlShellImage)

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var p *v1.Pod
	err = kom.Cluster(clusterID).WithContext(ctx).Resource(&v1.Pod{}).Name(podName).Namespace(ns).Get(&p).Error
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
