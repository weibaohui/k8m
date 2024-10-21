package default_cb

import (
	"context"
	"fmt"
	"reflect"

	"github.com/weibaohui/k8m/pkg/comm/kubectl"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
)

func List(ctx context.Context, k8s *kubectl.Kubectl) error {
	if klog.V(8).Enabled() {
		json := k8s.Statement.String()
		klog.V(8).Infof("DefaultCB List %s", json)
	}

	stmt := k8s.Statement
	gvr := stmt.GVR
	namespaced := stmt.Namespaced
	ns := stmt.Namespace
	ctx = stmt.Context

	// 使用反射获取 dest 的值
	destValue := reflect.ValueOf(stmt.Dest)

	// 确保 dest 是一个指向切片的指针
	if destValue.Kind() != reflect.Ptr || destValue.Elem().Kind() != reflect.Slice {
		// 处理错误：dest 不是指向切片的指针
		return fmt.Errorf("请传入数组类型")
	}
	// 获取切片的元素类型
	elemType := destValue.Elem().Type().Elem()

	var list *unstructured.UnstructuredList
	var err error

	if namespaced {
		list, err = stmt.DynamicClient.Resource(gvr).Namespace(ns).List(ctx, metav1.ListOptions{})
	} else {
		list, err = stmt.DynamicClient.Resource(gvr).List(ctx, metav1.ListOptions{})
	}
	if err != nil {
		return err
	}

	for _, item := range list.Items {
		obj := item.DeepCopy()
		stmt.RemoveManagedFields(obj)

		// 创建新的指向元素类型的指针
		newElemPtr := reflect.New(elemType)
		// unstructured 转换为原始目标类型
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, newElemPtr.Interface())
		// 将指针的值添加到切片中
		destValue.Elem().Set(reflect.Append(destValue.Elem(), newElemPtr.Elem()))

	}

	if err != nil {
		return err
	}
	return nil
}
