package default_cb

import (
	"context"
	"fmt"

	"github.com/weibaohui/k8m/pkg/comm/kubectl"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
)

func Patch(ctx context.Context, k8s *kubectl.Kubectl) error {
	if klog.V(8).Enabled() {
		json := k8s.Statement.String()
		klog.V(8).Infof("DefaultCB Patch %s", json)
	}

	stmt := k8s.Statement
	gvr := stmt.GVR
	namespaced := stmt.Namespaced
	ns := stmt.Namespace
	name := stmt.Name
	ctx = stmt.Context
	patchType := stmt.PatchType
	patchData := stmt.PatchData

	var res *unstructured.Unstructured
	var err error
	if name == "" {
		err = fmt.Errorf("patch对象必须指定名称")
		return err
	}
	if namespaced {
		res, err = stmt.DynamicClient.Resource(gvr).Namespace(ns).Patch(ctx, name, patchType, []byte(patchData), metav1.PatchOptions{})
	} else {
		res, err = stmt.DynamicClient.Resource(gvr).Patch(ctx, name, patchType, []byte(patchData), metav1.PatchOptions{})
	}
	if err != nil {
		return err
	}

	stmt.RemoveManagedFields(res)

	// 将 unstructured 转换回原始对象
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(res.Object, stmt.Dest)
	if err != nil {
		return err
	}

	return nil
}
