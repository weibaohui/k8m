package service

import (
	"fmt"
)

// PodLabels 定义结构体，用于存储Pod的标签信息
type PodLabels struct {
	ClusterName string            // 集群名称
	Namespace   string            // 命名空间
	PodName     string            // Pod名称
	Labels      map[string]string // 标签
}

// UpdatePodLabels 更新Pod的标签
func (p *podService) UpdatePodLabels(selectedCluster string, namespace string, podName string, labels map[string]string) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if p.podLabels == nil {
		p.podLabels = make(map[string][]*PodLabels)
	}
	// 查找是否已存在该Pod的标签
	var found bool
	if podList, ok := p.podLabels[selectedCluster]; ok {
		for _, pod := range podList {
			if pod.Namespace == namespace && pod.PodName == podName {
				pod.Labels = labels
				found = true
				break
			}
		}
	} else {
		p.podLabels[selectedCluster] = make([]*PodLabels, 0)
	}
	// 如果Pod不存在，则添加新Pod
	if !found {
		p.podLabels[selectedCluster] = append(p.podLabels[selectedCluster], &PodLabels{
			ClusterName: selectedCluster,
			Namespace:   namespace,
			PodName:     podName,
			Labels:      labels,
		})
	}
}

// DeletePodLabels 删除Pod的标签
func (p *podService) DeletePodLabels(selectedCluster string, namespace string, podName string) {
	p.lock.Lock()
	defer p.lock.Unlock()
	if podList, ok := p.podLabels[selectedCluster]; ok {
		for i, pod := range podList {
			if pod.Namespace == namespace && pod.PodName == podName {
				// 从切片中删除该Pod
				p.podLabels[selectedCluster] = append(podList[:i], podList[i+1:]...)
				break
			}
		}
	}
}

// GetPodLabels 获取Pod的标签
func (p *podService) GetPodLabels(selectedCluster string, namespace string, podName string) map[string]string {
	p.lock.RLock()
	defer p.lock.RUnlock()
	if podList, ok := p.podLabels[selectedCluster]; ok {
		for _, pod := range podList {
			if pod.Namespace == namespace && pod.PodName == podName {
				return pod.Labels
			}
		}
	}
	return nil
}

// GetAllPodLabels 获取所有Pod的标签
func (p *podService) GetAllPodLabels(selectedCluster string) map[string]map[string]string {
	p.lock.RLock()
	defer p.lock.RUnlock()
	// 创建一个新的map来返回，避免直接返回内部map
	result := make(map[string]map[string]string)
	if podList, ok := p.podLabels[selectedCluster]; ok {
		for _, pod := range podList {
			key := fmt.Sprintf("%s/%s", pod.Namespace, pod.PodName)
			labels := make(map[string]string)
			for k, v := range pod.Labels {
				labels[k] = v
			}
			result[key] = labels
		}
	}
	return result
}

// GetUniquePodLabels 获取所有Pod标签的唯一集合
func (p *podService) GetUniquePodLabels(selectedCluster string) map[string]string {
	p.lock.RLock()
	defer p.lock.RUnlock()
	// 创建一个新的map来存储唯一的标签
	uniqueLabels := make(map[string]string)
	// 遍历所有Pod的标签
	if podList, ok := p.podLabels[selectedCluster]; ok {
		for _, pod := range podList {
			// 将每个Pod的标签添加到唯一标签集合中
			for k, v := range pod.Labels {
				// 使用 k=v 作为键和值，以支持相同key但不同value的标签
				labelKey := fmt.Sprintf("%s=%s", k, v)
				uniqueLabels[labelKey] = labelKey
			}
		}
	}
	return uniqueLabels
}
