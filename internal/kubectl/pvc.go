package kubectl

import (
	"sort"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (k8s *Kubectl) ListPVC(ns string) ([]v1.PersistentVolumeClaim, error) {
	list, err := k8s.client.CoreV1().PersistentVolumeClaims(ns).List(k8s.Stmt.Context, metav1.ListOptions{})
	if err == nil && list != nil && list.Items != nil && len(list.Items) > 0 {
		// 按创建时间倒序排序 Pods 列表
		sort.Slice(list.Items, func(i, j int) bool {
			return list.Items[i].CreationTimestamp.Time.After(list.Items[j].CreationTimestamp.Time)
		})
		return list.Items, nil
	}
	return nil, err
}

func (k8s *Kubectl) GetPVC(ns, name string) (*v1.PersistentVolumeClaim, error) {
	pvc, err := k8s.client.CoreV1().PersistentVolumeClaims(ns).Get(k8s.Stmt.Context, name, metav1.GetOptions{})
	return pvc, err
}

func (k8s *Kubectl) CreatePVC(pvc *v1.PersistentVolumeClaim) (*v1.PersistentVolumeClaim, error) {
	pvc, err := k8s.client.CoreV1().PersistentVolumeClaims(pvc.Namespace).Create(k8s.Stmt.Context, pvc, metav1.CreateOptions{})
	return pvc, err
}
