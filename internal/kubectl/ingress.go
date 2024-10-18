package kubectl

import (
	"context"
	"sort"

	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k8s *Kubectl) ListIngress(ctx context.Context, ns string) ([]v1.Ingress, error) {
	list, err := k8s.client.NetworkingV1().Ingresses(ns).List(ctx, metav1.ListOptions{})
	if err == nil && list != nil && list.Items != nil && len(list.Items) > 0 {
		// 按创建时间倒序排序 Pods 列表
		sort.Slice(list.Items, func(i, j int) bool {
			return list.Items[i].CreationTimestamp.Time.After(list.Items[j].CreationTimestamp.Time)
		})
		return list.Items, nil
	}
	return nil, err
}

func (k8s *Kubectl) GetIngress(ctx context.Context, ns, name string) (*v1.Ingress, error) {
	Ingress, err := k8s.client.NetworkingV1().Ingresses(ns).Get(ctx, name, metav1.GetOptions{})
	return Ingress, err
}
func (k8s *Kubectl) RemoveIngress(ctx context.Context, ns, name string) error {
	err := k8s.client.NetworkingV1().Ingresses(ns).Delete(ctx, name, metav1.DeleteOptions{})
	return err
}
func (k8s *Kubectl) CreateIngress(ctx context.Context, Ingress *v1.Ingress) (*v1.Ingress, error) {
	Ingress, err := k8s.client.NetworkingV1().Ingresses(Ingress.Namespace).Create(ctx, Ingress, metav1.CreateOptions{})
	return Ingress, err
}
