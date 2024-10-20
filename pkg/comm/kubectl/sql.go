package kubectl

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/klog/v2"
)

func (k8s *Kubectl) SQLGet(obj runtime.Object) error {
	// 使用 scheme.Scheme.ObjectKinds() 获取 Kind
	gvks, _, err := scheme.Scheme.ObjectKinds(obj)
	if err != nil {
		return fmt.Errorf("error getting kind by scheme.Scheme.ObjectKinds : %v", err)
	}

	// 寻找kind
	var kind string
	typeAccessor, err := meta.TypeAccessor(obj)
	if err != nil {
		return fmt.Errorf("error getting typeMeta by meta.TypeAccessor(obj) : %v", err)
	}
	kind = typeAccessor.GetKind()

	if kind == "" {
		for _, gv := range gvks {
			if gv.Kind != "" {
				kind = gv.Kind
				break
			}
		}
	}

	// 获取元数据（比如Name和Namespace）
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return err
	}
	name := accessor.GetName()           // 获取资源的名称
	namespace := accessor.GetNamespace() // 获取资源的命名空间

	tx := k8s.getInstance()

	tx.Statement.SetGVKs(gvks).
		SetNamespace(namespace).
		SetName(name).
		SetType(Query).
		SetDest(obj)
	return tx.Callback().Query().Execute(tx.Statement.Context, tx)
}

func (k8s *Kubectl) sqlTest() {
	pod := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "",
		},
	}
	err := k8s.SQLGet(&pod)
	if err != nil {
		klog.Errorf("k8s.First(&pod) error :%v", err)
	}
}
