package kubectl

import (
	"context"
	"sort"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k8s *Kubectl) ListServiceAccount(ns string) ([]v1.ServiceAccount, error) {
	list, err := k8s.client.CoreV1().ServiceAccounts(ns).List(context.Background(), metav1.ListOptions{})
	if err == nil && list != nil && list.Items != nil && len(list.Items) > 0 {
		// 按创建时间倒序排序 Pods 列表
		sort.Slice(list.Items, func(i, j int) bool {
			return list.Items[i].CreationTimestamp.Time.After(list.Items[j].CreationTimestamp.Time)
		})
		return list.Items, nil
	}
	return nil, err
}

func (k8s *Kubectl) GetServiceAccount(ns, name string) (*v1.ServiceAccount, error) {
	ServiceAccount, err := k8s.client.CoreV1().ServiceAccounts(ns).Get(context.Background(), name, metav1.GetOptions{})
	return ServiceAccount, err
}
func (k8s *Kubectl) RemoveServiceAccount(ns, name string) error {
	err := k8s.client.CoreV1().ServiceAccounts(ns).Delete(context.Background(), name, metav1.DeleteOptions{})
	return err
}
func (k8s *Kubectl) CreateServiceAccount(sa *v1.ServiceAccount) (*v1.ServiceAccount, error) {
	sa, err := k8s.client.CoreV1().ServiceAccounts(sa.Namespace).Create(context.Background(), sa, metav1.CreateOptions{})
	return sa, err
}

func (k8s *Kubectl) UpdateServiceAccount(sa *v1.ServiceAccount) (*v1.ServiceAccount, error) {
	sa, err := k8s.client.CoreV1().ServiceAccounts(sa.Namespace).Update(context.Background(), sa, metav1.UpdateOptions{})
	return sa, err
}
