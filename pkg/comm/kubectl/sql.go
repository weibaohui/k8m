package kubectl

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

func (k8s *Kubectl) WithContext(ctx context.Context) *Kubectl {
	tx := k8s.getInstance()
	tx.Statement.Context = ctx
	return tx
}
func (k8s *Kubectl) Resource(obj runtime.Object) *Kubectl {
	tx := k8s.getInstance()
	tx.Statement.ParseFromRuntimeObj(obj)
	return tx
}
func (k8s *Kubectl) Namespace(ns string) *Kubectl {
	tx := k8s.getInstance()
	tx.Statement.Namespace = ns
	return tx
}
func (k8s *Kubectl) Name(ns string) *Kubectl {
	tx := k8s.getInstance()
	tx.Statement.Name = ns
	return tx
}

func (k8s *Kubectl) CRD(group string, version string, kind string) *Kubectl {
	gvk := schema.GroupVersionKind{
		Group:   group,
		Version: version,
		Kind:    kind,
	}
	k8s.Statement.ParseGVKs([]schema.GroupVersionKind{
		gvk,
	})

	return k8s
}
func (k8s *Kubectl) Unstructured() *Kubectl {
	tx := k8s.getInstance()
	tx.Statement.Unstructured = true
	return tx
}
func (k8s *Kubectl) Get(obj runtime.Object) error {
	tx := k8s.getInstance()

	tx.Statement.SetType(Query)
	tx.Statement.Dest = obj
	return tx.Callback().Query().Execute(tx.Statement.Context, tx)
}
func (k8s *Kubectl) Fill(m *unstructured.Unstructured) error {
	tx := k8s.getInstance()
	err := tx.Callback().Query().Execute(tx.Statement.Context, tx)
	if err != nil {
		return err
	}
	// 确保将数据填充到传入的 m 中
	if dest, ok := tx.Statement.Dest.(*unstructured.Unstructured); ok {
		*m = *dest
	}
	return nil
}

func (k8s *Kubectl) sqlTest() {
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "",
		},
	}
	err := k8s.Resource(&pod).
		Namespace("default").Name("").Get(&pod)
	if err != nil {
		klog.Errorf("k8s.First(&pod) error :%v", err)
	}
}
