package kubectl

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
func (k8s *Kubectl) Get(obj runtime.Object) error {
	tx := k8s.getInstance()

	tx.Statement.SetType(Query)
	tx.Statement.Dest = obj
	return tx.Callback().Query().Execute(tx.Statement.Context, tx)
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
