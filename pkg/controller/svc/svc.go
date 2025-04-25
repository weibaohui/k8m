package svc

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// Create 创建Service接口
func Create(c *gin.Context) {
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var req struct {
		Metadata struct {
			Namespace string            `json:"namespace"`
			Name      string            `json:"name"`
			Labels    map[string]string `json:"labels"`
		} `json:"metadata"`
		Spec struct {
			Type  corev1.ServiceType `json:"type"`
			Ports []struct {
				Name       string `json:"name"`
				Port       int32  `json:"port"`
				TargetPort int32  `json:"targetPort"`
				NodePort   int32  `json:"nodePort"`
				Protocol   string `json:"protocol"`
			} `json:"ports"`
			Selector map[string]string `json:"selector"`
		} `json:"spec"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	// 判断是否存在同名Service
	var existingService corev1.Service
	err = kom.Cluster(selectedCluster).WithContext(ctx).Resource(&corev1.Service{}).Name(req.Metadata.Name).Namespace(req.Metadata.Namespace).Get(&existingService).Error
	if err == nil {
		amis.WriteJsonError(c, fmt.Errorf("Service %s 已存在", req.Metadata.Name))
		return
	}
	// 转换端口配置
	var k8sPorts []corev1.ServicePort
	for _, p := range req.Spec.Ports {
		port := corev1.ServicePort{
			Port:       p.Port,
			Name:       p.Name,
			TargetPort: intstr.FromInt32(p.TargetPort),
			Protocol:   corev1.Protocol(p.Protocol),
		}
		// 如果前端提供了协议
		if p.Protocol != "" {
			port.Protocol = corev1.Protocol(p.Protocol)
		}
		// 如果是NodePort类型且指定了nodePort
		if req.Spec.Type == corev1.ServiceTypeNodePort && p.NodePort > 0 {
			port.NodePort = p.NodePort
		}
		k8sPorts = append(k8sPorts, port)
	}

	// 创建Service对象
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Metadata.Name,
			Namespace: req.Metadata.Namespace,
			Labels:    req.Metadata.Labels,
		},
		Spec: corev1.ServiceSpec{
			Type:     req.Spec.Type,
			Ports:    k8sPorts,
			Selector: req.Spec.Selector,
		},
	}

	// 调用Kubernetes API
	err = kom.Cluster(selectedCluster).
		WithContext(ctx).
		Resource(&corev1.Service{}).
		Namespace(req.Metadata.Namespace).
		Create(service).Error

	amis.WriteJsonErrorOrOK(c, err)
}
