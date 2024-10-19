package kubectl

import (
	"context"
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
	gvk, _, err := scheme.Scheme.ObjectKinds(obj)
	if err != nil {
		fmt.Println("Error getting kind:", err)
	}

	var kind string
	// 打印 Kind 信息
	for _, gv := range gvk {
		kind = gv.Kind
		fmt.Printf("Group: %s, Version: %s, Kind: %s\n", gv.Group, gv.Version, gv.Kind)
	}

	// 获取元数据（比如Name和Namespace）
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return err
	}

	name := accessor.GetName()           // 获取资源的名称
	namespace := accessor.GetNamespace() // 获取资源的命名空间

	resource, err := kubectl.GetResource(context.TODO(), kind, namespace, name)
	if err != nil {
		return err
	}

	// 将 unstructured 转换回原始对象
	return runtime.DefaultUnstructuredConverter.FromUnstructured(resource.Object, obj)
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
