package default_cb

import (
	"context"

	"github.com/weibaohui/k8m/pkg/comm/kubectl"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
)

func Get(ctx context.Context, k8s *kubectl.Kubectl) error {
	if klog.V(8).Enabled() {
		json := k8s.Statement.String()
		klog.V(8).Infof("DefaultCB Get %s", json)
	}

	stmt := k8s.Statement
	gvr := stmt.GVR
	namespaced := stmt.Namespaced
	ns := stmt.Namespace
	name := stmt.Name
	ctx = stmt.Context
	var res *unstructured.Unstructured
	var err error

	if namespaced {
		res, err = stmt.DynamicClient.Resource(gvr).Namespace(ns).Get(ctx, name, metav1.GetOptions{})
	} else {
		res, err = stmt.DynamicClient.Resource(gvr).Get(ctx, name, metav1.GetOptions{})
	}
	if err != nil {
		return err
	}

	// 将 unstructured 转换回原始对象
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(res.Object, stmt.Dest)
	if err != nil {
		return err
	}
	return nil
}
