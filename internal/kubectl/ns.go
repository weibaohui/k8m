package kubectl

import (
	"context"
	"sort"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k8s *Kubectl) ListNamespace() ([]v1.Namespace, error) {
	list, err := k8s.client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err == nil && list != nil && list.Items != nil && len(list.Items) > 0 {
		// 按创建时间倒序排序 Pods 列表
		sort.Slice(list.Items, func(i, j int) bool {
			return list.Items[i].CreationTimestamp.Time.After(list.Items[j].CreationTimestamp.Time)
		})
		return list.Items, nil
	}
	return nil, err
}

func (k8s *Kubectl) GetNamespace(name string) (*v1.Namespace, error) {
	Namespace, err := k8s.client.CoreV1().Namespaces().Get(context.TODO(), name, metav1.GetOptions{})
	return Namespace, err
}

func (k8s *Kubectl) RemoveNamespace(name string) error {
	err := k8s.client.CoreV1().Namespaces().Delete(context.TODO(), name, metav1.DeleteOptions{})
	return err
}
func (k8s *Kubectl) CreateNamespace(ns *v1.Namespace) (*v1.Namespace, error) {
	ns, err := k8s.client.CoreV1().Namespaces().Create(context.TODO(), ns, metav1.CreateOptions{})
	return ns, err
}
