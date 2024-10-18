package kubectl

import (
	"context"
	"sort"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k8s *Kubectl) ListService(ctx context.Context, ns string) ([]v1.Service, error) {
	list, err := k8s.client.CoreV1().Services(ns).List(ctx, metav1.ListOptions{})
	if err == nil && list != nil && list.Items != nil && len(list.Items) > 0 {
		// 按创建时间倒序排序 Pods 列表
		sort.Slice(list.Items, func(i, j int) bool {
			return list.Items[i].CreationTimestamp.Time.After(list.Items[j].CreationTimestamp.Time)
		})
		return list.Items, nil
	}
	return nil, err
}

func (k8s *Kubectl) GetService(ctx context.Context, ns, name string) (*v1.Service, error) {
	Service, err := k8s.client.CoreV1().Services(ns).Get(ctx, name, metav1.GetOptions{})
	return Service, err
}
func (k8s *Kubectl) RemoveService(ctx context.Context, ns, name string) error {
	err := k8s.client.CoreV1().Services(ns).Delete(ctx, name, metav1.DeleteOptions{})
	return err
}
func (k8s *Kubectl) CreateService(ctx context.Context, svc *v1.Service) (*v1.Service, error) {
	svc, err := k8s.client.CoreV1().Services(svc.Namespace).Create(ctx, svc, metav1.CreateOptions{})
	return svc, err
}
