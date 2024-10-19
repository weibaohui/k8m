package kubectl

import (
	"context"
	"sort"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k8s *Kubectl) ListNode(ctx context.Context) ([]v1.Node, error) {
	list, err := k8s.client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err == nil && list != nil && list.Items != nil && len(list.Items) > 0 {
		sort.Slice(list.Items, func(i, j int) bool {
			return list.Items[i].CreationTimestamp.Time.After(list.Items[j].CreationTimestamp.Time)
		})
		return list.Items, nil
	}
	return nil, err
}

func (k8s *Kubectl) GetNode(ctx context.Context, name string) (*v1.Node, error) {
	node, err := k8s.client.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
	return node, err
}
