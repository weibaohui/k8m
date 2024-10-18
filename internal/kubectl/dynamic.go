package kubectl

import (
	"context"
	"fmt"
	"sort"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
)

// ListOption Functional options for ListResources
type ListOption func(*metav1.ListOptions)

// WithLabelSelector 设置 LabelSelector
func WithLabelSelector(labelSelector string) ListOption {
	return func(lo *metav1.ListOptions) {
		lo.LabelSelector = labelSelector
	}
}

// WithFieldSelector 设置 FieldSelector
func WithFieldSelector(fieldSelector string) ListOption {
	return func(lo *metav1.ListOptions) {
		lo.FieldSelector = fieldSelector
	}
}

func (k8s *Kubectl) ListResources(kind string, ns string, opts ...ListOption) ([]unstructured.Unstructured, error) {
	gvr, namespaced := k8s.GetGVR(kind)
	if gvr.Empty() {
		return nil, fmt.Errorf("不支持的资源类型: %s", kind)
	}

	listOptions := metav1.ListOptions{}
	for _, opt := range opts {
		opt(&listOptions)
	}

	var err error

	k8s.Stmt.SetGVR(gvr).SetNamespace(ns).
		SetName("").
		SetKind(kind).
		SetNamespaced(namespaced).
		SetType(Query).
		SetListOptions(&listOptions)
	err = k8s.Callback().Query().Execute(k8s)
	if err != nil {
		return nil, err
	}

	var list *unstructured.UnstructuredList
	if namespaced {
		list, err = k8s.dynamicClient.Resource(gvr).Namespace(ns).List(context.TODO(), listOptions)
	} else {
		list, err = k8s.dynamicClient.Resource(gvr).List(context.TODO(), listOptions)
	}
	if err != nil {
		return nil, err
	}
	var resources []unstructured.Unstructured
	for _, item := range list.Items {
		obj := item.DeepCopy()
		removeManagedFields(obj)
		resources = append(resources, *obj)
	}

	return sortByCreationTime(resources), nil
}
func (k8s *Kubectl) GetResource(kind string, ns, name string) (*unstructured.Unstructured, error) {
	gvr, namespaced := k8s.GetGVR(kind)
	if gvr.Empty() {
		return nil, fmt.Errorf("不支持的资源类型: %s", kind)
	}
	var res *unstructured.Unstructured
	var err error

	k8s.Stmt.SetGVR(gvr).SetNamespace(ns).
		SetName(name).
		SetKind(kind).
		SetNamespaced(namespaced).
		SetType(Query)
	err = k8s.Callback().Query().Execute(k8s)
	if err != nil {
		return nil, err
	}

	if namespaced {
		res, err = k8s.dynamicClient.Resource(gvr).Namespace(ns).Get(context.TODO(), name, metav1.GetOptions{})
	} else {
		res, err = k8s.dynamicClient.Resource(gvr).Get(context.TODO(), name, metav1.GetOptions{})
	}
	if err != nil {
		return nil, err
	}

	removeManagedFields(res)
	return res, nil
}
func (k8s *Kubectl) CreateResource(kind string, ns string, resource *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	gvr, namespaced := k8s.GetGVR(kind)
	if gvr.Empty() {
		return nil, fmt.Errorf("不支持的资源类型: %s", kind)
	}
	var createdResource *unstructured.Unstructured
	var err error
	k8s.Stmt.SetGVR(gvr).
		SetNamespace(resource.GetNamespace()).
		SetName(resource.GetName()).
		SetKind(kind).
		SetNamespaced(namespaced).
		SetType(Create)
	err = k8s.Callback().Create().Execute(k8s)
	if err != nil {
		return nil, err
	}
	if namespaced {
		createdResource, err = k8s.dynamicClient.Resource(gvr).Namespace(ns).Create(context.TODO(), resource, metav1.CreateOptions{})
	} else {
		createdResource, err = k8s.dynamicClient.Resource(gvr).Create(context.TODO(), resource, metav1.CreateOptions{})
	}
	if err != nil {
		return nil, err
	}

	removeManagedFields(createdResource)
	return createdResource, nil
}

func (k8s *Kubectl) UpdateResource(kind string, ns string, resource *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	gvr, namespaced := k8s.GetGVR(kind)
	if gvr.Empty() {
		return nil, fmt.Errorf("不支持的资源类型: %s", kind)
	}
	var updatedResource *unstructured.Unstructured
	var err error
	k8s.Stmt.SetGVR(gvr).
		SetNamespace(resource.GetNamespace()).
		SetName(resource.GetName()).
		SetKind(kind).
		SetNamespaced(namespaced).
		SetType(Update)
	err = k8s.Callback().Update().Execute(k8s)
	if err != nil {
		return nil, err
	}
	if namespaced {
		updatedResource, err = k8s.dynamicClient.Resource(gvr).Namespace(ns).Update(context.TODO(), resource, metav1.UpdateOptions{})
	} else {
		updatedResource, err = k8s.dynamicClient.Resource(gvr).Update(context.TODO(), resource, metav1.UpdateOptions{})
	}

	if err != nil {
		return nil, fmt.Errorf("无法更新资源: %v", err)
	}
	removeManagedFields(updatedResource)
	return updatedResource, nil
}

func (k8s *Kubectl) DeleteResource(kind string, ns, name string) error {
	gvr, namespaced := k8s.GetGVR(kind)
	if gvr.Empty() {
		return fmt.Errorf("不支持的资源类型: %s", kind)
	}

	var err error
	k8s.Stmt.SetGVR(gvr).
		SetNamespace(ns).
		SetName(name).
		SetKind(kind).
		SetNamespaced(namespaced).
		SetType(Delete)
	err = k8s.Callback().Delete().Execute(k8s)
	if err != nil {
		return err
	}

	if namespaced {
		return k8s.dynamicClient.Resource(gvr).Namespace(ns).Delete(context.TODO(), name, metav1.DeleteOptions{})
	} else {
		return k8s.dynamicClient.Resource(gvr).Delete(context.TODO(), name, metav1.DeleteOptions{})
	}
}

func (k8s *Kubectl) PatchResource(kind string, ns, name string, patchType types.PatchType, patchData []byte) (*unstructured.Unstructured, error) {
	gvr, namespaced := k8s.GetGVR(kind)
	if gvr.Empty() {
		return nil, fmt.Errorf("不支持的资源类型: %s", kind)
	}
	var obj *unstructured.Unstructured
	var err error
	k8s.Stmt.SetGVR(gvr).
		SetNamespace(ns).
		SetName(name).
		SetKind(kind).
		SetNamespaced(namespaced).
		SetType(Patch)
	err = k8s.Callback().Update().Execute(k8s)
	if err != nil {
		return nil, err
	}

	if namespaced {
		obj, err = k8s.dynamicClient.Resource(gvr).Namespace(ns).Patch(context.TODO(), name, patchType, patchData, metav1.PatchOptions{})
	} else {
		obj, err = k8s.dynamicClient.Resource(gvr).Patch(context.TODO(), name, patchType, patchData, metav1.PatchOptions{})
	}
	if err != nil {
		return nil, err
	}

	removeManagedFields(obj)
	return obj, nil
}

// splitYAML 按 "---" 分割多文档 YAML
func splitYAML(yamlStr string) []string {
	return strings.Split(yamlStr, "\n---\n")
}

// removeManagedFields 删除 unstructured.Unstructured 对象中的 metadata.managedFields 字段
func removeManagedFields(obj *unstructured.Unstructured) {
	// 获取 metadata
	metadata, found, err := unstructured.NestedMap(obj.Object, "metadata")
	if err != nil || !found {
		return
	}

	// 删除 managedFields
	delete(metadata, "managedFields")

	// 更新 metadata
	err = unstructured.SetNestedMap(obj.Object, metadata, "metadata")
	if err != nil {
		return
	}
}

// sortByCreationTime 按创建时间排序资源
func sortByCreationTime(items []unstructured.Unstructured) []unstructured.Unstructured {
	sort.Slice(items, func(i, j int) bool {
		ti := items[i].GetCreationTimestamp()
		tj := items[j].GetCreationTimestamp()
		return ti.After(tj.Time)
	})
	return items
}
