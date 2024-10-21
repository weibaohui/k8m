package kubectl

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/weibaohui/k8m/pkg/comm/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
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

func (k8s *Kubectl) ListResources(ctx context.Context, kind string, ns string, opts ...ListOption) ([]unstructured.Unstructured, error) {
	gvr, namespaced := k8s.GetGVR(kind)
	if gvr.Empty() {
		return nil, fmt.Errorf("不支持的资源类型: %s", kind)
	}

	listOptions := metav1.ListOptions{}
	for _, opt := range opts {
		opt(&listOptions)
	}

	var err error

	var list *unstructured.UnstructuredList
	if namespaced {
		list, err = k8s.dynamicClient.Resource(gvr).Namespace(ns).List(ctx, listOptions)
	} else {
		list, err = k8s.dynamicClient.Resource(gvr).List(ctx, listOptions)
	}
	if err != nil {
		return nil, err
	}
	var resources []unstructured.Unstructured
	for _, item := range list.Items {
		obj := item.DeepCopy()
		k8s.RemoveManagedFields(obj)
		resources = append(resources, *obj)
	}

	return sortByCreationTime(resources), nil
}
func (k8s *Kubectl) GetResource(ctx context.Context, kind string, ns, name string) (*unstructured.Unstructured, error) {
	gvr, namespaced := k8s.GetGVR(kind)
	gvrString := utils.ToJSON(gvr)
	klog.V(8).Infof("(k8s *Kubectl) GetResource GVR %s", gvrString)
	if gvr.Empty() {
		return nil, fmt.Errorf("不支持的资源类型: %s", kind)
	}
	var res *unstructured.Unstructured
	var err error

	if namespaced {
		res, err = k8s.dynamicClient.Resource(gvr).Namespace(ns).Get(ctx, name, metav1.GetOptions{})
	} else {
		res, err = k8s.dynamicClient.Resource(gvr).Get(ctx, name, metav1.GetOptions{})
	}
	if err != nil {
		return nil, err
	}

	k8s.RemoveManagedFields(res)
	return res, nil
}
func (k8s *Kubectl) CreateResource(ctx context.Context, kind string, ns string, resource *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	gvr, namespaced := k8s.GetGVR(kind)
	if gvr.Empty() {
		return nil, fmt.Errorf("不支持的资源类型: %s", kind)
	}
	var createdResource *unstructured.Unstructured
	var err error

	if namespaced {
		createdResource, err = k8s.dynamicClient.Resource(gvr).Namespace(ns).Create(ctx, resource, metav1.CreateOptions{})
	} else {
		createdResource, err = k8s.dynamicClient.Resource(gvr).Create(ctx, resource, metav1.CreateOptions{})
	}
	if err != nil {
		return nil, err
	}

	k8s.RemoveManagedFields(createdResource)
	return createdResource, nil
}

func (k8s *Kubectl) UpdateResource(ctx context.Context, kind string, ns string, resource *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	gvr, namespaced := k8s.GetGVR(kind)
	if gvr.Empty() {
		return nil, fmt.Errorf("不支持的资源类型: %s", kind)
	}
	var updatedResource *unstructured.Unstructured
	var err error

	if namespaced {
		updatedResource, err = k8s.dynamicClient.Resource(gvr).Namespace(ns).Update(ctx, resource, metav1.UpdateOptions{})
	} else {
		updatedResource, err = k8s.dynamicClient.Resource(gvr).Update(ctx, resource, metav1.UpdateOptions{})
	}

	if err != nil {
		return nil, fmt.Errorf("无法更新资源: %v", err)
	}
	k8s.RemoveManagedFields(updatedResource)
	return updatedResource, nil
}

func (k8s *Kubectl) DeleteResource(ctx context.Context, kind string, ns, name string) error {
	gvr, namespaced := k8s.GetGVR(kind)
	if gvr.Empty() {
		return fmt.Errorf("不支持的资源类型: %s", kind)
	}

	if namespaced {
		return k8s.dynamicClient.Resource(gvr).Namespace(ns).Delete(ctx, name, metav1.DeleteOptions{})
	} else {
		return k8s.dynamicClient.Resource(gvr).Delete(ctx, name, metav1.DeleteOptions{})
	}
}

func (k8s *Kubectl) PatchResource(ctx context.Context, kind string, ns, name string, patchType types.PatchType, patchData []byte) (*unstructured.Unstructured, error) {
	gvr, namespaced := k8s.GetGVR(kind)
	if gvr.Empty() {
		return nil, fmt.Errorf("不支持的资源类型: %s", kind)
	}
	var obj *unstructured.Unstructured
	var err error

	if namespaced {
		obj, err = k8s.dynamicClient.Resource(gvr).Namespace(ns).Patch(ctx, name, patchType, patchData, metav1.PatchOptions{})
	} else {
		obj, err = k8s.dynamicClient.Resource(gvr).Patch(ctx, name, patchType, patchData, metav1.PatchOptions{})
	}
	if err != nil {
		return nil, err
	}

	k8s.RemoveManagedFields(obj)
	return obj, nil
}

// splitYAML 按 "---" 分割多文档 YAML
func splitYAML(yamlStr string) []string {
	return strings.Split(yamlStr, "\n---\n")
}

// RemoveManagedFields 删除 unstructured.Unstructured 对象中的 metadata.managedFields 字段
func (k8s *Kubectl) RemoveManagedFields(obj *unstructured.Unstructured) {
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
