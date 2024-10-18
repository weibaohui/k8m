package kubectl

import (
	"sort"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k8s *Kubectl) ListServiceAccount(ns string) ([]v1.ServiceAccount, error) {
	list, err := k8s.client.CoreV1().ServiceAccounts(ns).List(k8s.Stmt.Context, metav1.ListOptions{})
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
	ServiceAccount, err := k8s.client.CoreV1().ServiceAccounts(ns).Get(k8s.Stmt.Context, name, metav1.GetOptions{})
	return ServiceAccount, err
}
func (k8s *Kubectl) RemoveServiceAccount(ns, name string) error {
	err := k8s.client.CoreV1().ServiceAccounts(ns).Delete(k8s.Stmt.Context, name, metav1.DeleteOptions{})
	return err
}
func (k8s *Kubectl) CreateServiceAccount(sa *v1.ServiceAccount) (*v1.ServiceAccount, error) {
	sa, err := k8s.client.CoreV1().ServiceAccounts(sa.Namespace).Create(k8s.Stmt.Context, sa, metav1.CreateOptions{})
	return sa, err
}

func (k8s *Kubectl) UpdateServiceAccount(sa *v1.ServiceAccount) (*v1.ServiceAccount, error) {
	sa, err := k8s.client.CoreV1().ServiceAccounts(sa.Namespace).Update(k8s.Stmt.Context, sa, metav1.UpdateOptions{})
	return sa, err
}
