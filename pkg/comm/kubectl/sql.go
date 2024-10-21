package kubectl

import (
	"context"
	"fmt"

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

func (k8s *Kubectl) Get(dest interface{}) *Kubectl {
	tx := k8s.getInstance()
	tx.Statement.SetType(Get)
	// 设置目标对象为 obj 的指针
	tx.Statement.Dest = dest
	tx.Error = tx.Callback().Get().Execute(tx.Statement.Context, tx)
	return tx
}
func (k8s *Kubectl) List(dest interface{}) *Kubectl {
	tx := k8s.getInstance()
	tx.Statement.SetType(List)
	tx.Statement.Dest = dest
	tx.Error = tx.Callback().List().Execute(tx.Statement.Context, tx)
	return tx
}
func (k8s *Kubectl) Fill(m *unstructured.Unstructured) *Kubectl {
	tx := k8s.getInstance()
	if tx.Statement.Dest == nil {
		tx.Error = fmt.Errorf("请先执行Get()、List()等方法")
	}
	// 确保将数据填充到传入的 m 中
	if dest, ok := tx.Statement.Dest.(*unstructured.Unstructured); ok {
		*m = *dest
	}
	return k8s
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
