package kubectl

import (
	"context"
	"sort"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k8s *Kubectl) ListSecret(ns string) ([]v1.Secret, error) {
	list, err := k8s.client.CoreV1().Secrets(ns).List(context.TODO(), metav1.ListOptions{})
	if err == nil && list != nil && list.Items != nil && len(list.Items) > 0 {
		// 按创建时间倒序排序 Pods 列表
		sort.Slice(list.Items, func(i, j int) bool {
			return list.Items[i].CreationTimestamp.Time.After(list.Items[j].CreationTimestamp.Time)
		})
		return list.Items, nil
	}
	return nil, err
}

func (k8s *Kubectl) GetSecret(ns, name string) (*v1.Secret, error) {
	Secret, err := k8s.client.CoreV1().Secrets(ns).Get(context.TODO(), name, metav1.GetOptions{})
	return Secret, err
}
func (k8s *Kubectl) CreateSecret(secret *v1.Secret) (*v1.Secret, error) {
	secret, err := k8s.client.CoreV1().Secrets(secret.Namespace).Create(context.TODO(), secret, metav1.CreateOptions{})
	return secret, err
}
