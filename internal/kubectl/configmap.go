package kubectl

import (
	"context"
	"sort"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k8s *Kubectl) ListConfigMap(ns string) ([]v1.ConfigMap, error) {
	list, err := k8s.client.CoreV1().ConfigMaps(ns).List(context.TODO(), metav1.ListOptions{})
	if err == nil && list != nil && list.Items != nil && len(list.Items) > 0 {
		// 按创建时间倒序排序 Pods 列表
		sort.Slice(list.Items, func(i, j int) bool {
			return list.Items[i].CreationTimestamp.Time.After(list.Items[j].CreationTimestamp.Time)
		})
		return list.Items, nil
	}
	return nil, err
}

func (k8s *Kubectl) GetConfigMap(ns, name string) (*v1.ConfigMap, error) {
	ConfigMap, err := k8s.client.CoreV1().ConfigMaps(ns).Get(context.TODO(), name, metav1.GetOptions{})
	return ConfigMap, err
}
func (k8s *Kubectl) RemoveConfigMap(ns, name string) error {
	err := k8s.client.CoreV1().ConfigMaps(ns).Delete(context.TODO(), name, metav1.DeleteOptions{})
	return err
}
func (k8s *Kubectl) CreateConfigMap(cm *v1.ConfigMap) (*v1.ConfigMap, error) {
	cm, err := k8s.client.CoreV1().ConfigMaps(cm.Namespace).Create(context.TODO(), cm, metav1.CreateOptions{})
	return cm, err
}
