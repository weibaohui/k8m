package utils

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// LabelsManager 结构体，包含共享标签
type LabelsManager struct {
	Labels map[string]string
}

// NewLabelsManager 构造函数，初始化并返回一个 LabelsManager
func NewLabelsManager(labels map[string]string) *LabelsManager {
	return &LabelsManager{
		Labels: labels,
	}
}

// AddLabels 给任意 Kubernetes 资源对象添加共享标签
func (lm *LabelsManager) AddLabels(meta *metav1.ObjectMeta) {
	if meta.Labels == nil {
		meta.Labels = make(map[string]string)
	}
	for k, v := range lm.Labels {
		meta.Labels[k] = v
	}
}

// AddCustomLabel 动态添加用户指定的标签
func (lm *LabelsManager) AddCustomLabel(meta *metav1.ObjectMeta, key, value string) {
	if meta.Labels == nil {
		meta.Labels = make(map[string]string)
	}
	meta.Labels[key] = value
}
