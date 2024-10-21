package default_cb

import (
	"context"
	"fmt"

	"github.com/weibaohui/k8m/pkg/comm/kubectl"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

func Delete(ctx context.Context, k8s *kubectl.Kubectl) error {
	if klog.V(8).Enabled() {
		json := k8s.Statement.String()
		klog.V(8).Infof("DefaultCB Delete %s", json)
	}

	stmt := k8s.Statement
	gvr := stmt.GVR
	namespaced := stmt.Namespaced
	ns := stmt.Namespace
	name := stmt.Name
	ctx = stmt.Context
	var err error
	if name == "" {
		err = fmt.Errorf("删除对象必须指定名称")
		return err
	}

	if namespaced {
		err = stmt.DynamicClient.Resource(gvr).Namespace(ns).Delete(ctx, name, metav1.DeleteOptions{})
	} else {
		err = stmt.DynamicClient.Resource(gvr).Delete(ctx, name, metav1.DeleteOptions{})
	}

	if err != nil {
		return err
	}

	return nil
}
