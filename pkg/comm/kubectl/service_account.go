package kubectl

import (
	"context"
	"sort"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k8s *Kubectl) ListServiceAccount(ctx context.Context, ns string) ([]v1.ServiceAccount, error) {
	list, err := k8s.client.CoreV1().ServiceAccounts(ns).List(ctx, metav1.ListOptions{})
	if err == nil && list != nil && list.Items != nil && len(list.Items) > 0 {
		// 按创建时间倒序排序 Pods 列表
		sort.Slice(list.Items, func(i, j int) bool {
			return list.Items[i].CreationTimestamp.Time.After(list.Items[j].CreationTimestamp.Time)
		})
		return list.Items, nil
	}
	return nil, err
}

func (k8s *Kubectl) GetServiceAccount(ctx context.Context, ns, name string) (*v1.ServiceAccount, error) {
	ServiceAccount, err := k8s.client.CoreV1().ServiceAccounts(ns).Get(ctx, name, metav1.GetOptions{})
	return ServiceAccount, err
}
func (k8s *Kubectl) RemoveServiceAccount(ctx context.Context, ns, name string) error {
	err := k8s.client.CoreV1().ServiceAccounts(ns).Delete(ctx, name, metav1.DeleteOptions{})
	return err
}
func (k8s *Kubectl) CreateServiceAccount(ctx context.Context, sa *v1.ServiceAccount) (*v1.ServiceAccount, error) {
	sa, err := k8s.client.CoreV1().ServiceAccounts(sa.Namespace).Create(ctx, sa, metav1.CreateOptions{})
	return sa, err
}

func (k8s *Kubectl) UpdateServiceAccount(ctx context.Context, sa *v1.ServiceAccount) (*v1.ServiceAccount, error) {
	sa, err := k8s.client.CoreV1().ServiceAccounts(sa.Namespace).Update(ctx, sa, metav1.UpdateOptions{})
	return sa, err
}
