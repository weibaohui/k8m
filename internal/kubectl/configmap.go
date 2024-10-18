package kubectl

import (
	"context"
	"sort"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k8s *Kubectl) ListConfigMap(ctx context.Context, ns string) ([]v1.ConfigMap, error) {
	list, err := k8s.client.CoreV1().ConfigMaps(ns).List(ctx, metav1.ListOptions{})
	if err == nil && list != nil && list.Items != nil && len(list.Items) > 0 {
		// 按创建时间倒序排序 Pods 列表
		sort.Slice(list.Items, func(i, j int) bool {
			return list.Items[i].CreationTimestamp.Time.After(list.Items[j].CreationTimestamp.Time)
		})
		return list.Items, nil
	}
	return nil, err
}

func (k8s *Kubectl) GetConfigMap(ctx context.Context, ns, name string) (*v1.ConfigMap, error) {
	ConfigMap, err := k8s.client.CoreV1().ConfigMaps(ns).Get(ctx, name, metav1.GetOptions{})
	return ConfigMap, err
}
func (k8s *Kubectl) RemoveConfigMap(ctx context.Context, ns, name string) error {
	err := k8s.client.CoreV1().ConfigMaps(ns).Delete(ctx, name, metav1.DeleteOptions{})
	return err
}
func (k8s *Kubectl) CreateConfigMap(ctx context.Context, cm *v1.ConfigMap) (*v1.ConfigMap, error) {
	cm, err := k8s.client.CoreV1().ConfigMaps(cm.Namespace).Create(ctx, cm, metav1.CreateOptions{})
	return cm, err
}
