package kubectl

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

func (k8s *Kubectl) ListResourcesDynamic(gvr schema.GroupVersionResource, isNamespaced bool, ns string, opts ...ListOption) ([]unstructured.Unstructured, error) {
	if gvr.Empty() {
		return nil, fmt.Errorf("GroupVersionResource is empty")
	}

	listOptions := metav1.ListOptions{}
	for _, opt := range opts {
		opt(&listOptions)
	}
	var list *unstructured.UnstructuredList
	var err error
	if isNamespaced {
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
func (k8s *Kubectl) GetResourceDynamic(gvr schema.GroupVersionResource, isNamespaced bool, ns, name string) (*unstructured.Unstructured, error) {
	if gvr.Empty() {
		return nil, fmt.Errorf("GroupVersionResource is empty")
	}
	var obj *unstructured.Unstructured
	var err error
	if isNamespaced {
		obj, err = k8s.dynamicClient.Resource(gvr).Namespace(ns).Get(context.TODO(), name, metav1.GetOptions{})
	} else {
		obj, err = k8s.dynamicClient.Resource(gvr).Get(context.TODO(), name, metav1.GetOptions{})
	}
	if err != nil {
		return nil, err
	}

	removeManagedFields(obj)
	return obj, nil
}
func (k8s *Kubectl) CreateResourceDynamic(gvr schema.GroupVersionResource, isNamespaced bool, resource *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	if gvr.Empty() {
		return nil, fmt.Errorf("GroupVersionResource is empty")
	}
	var createdResource *unstructured.Unstructured
	var err error
	if isNamespaced {
		createdResource, err = k8s.dynamicClient.Resource(gvr).Namespace(resource.GetNamespace()).Create(context.TODO(), resource, metav1.CreateOptions{})
	} else {
		createdResource, err = k8s.dynamicClient.Resource(gvr).Create(context.TODO(), resource, metav1.CreateOptions{})
	}
	if err != nil {
		return nil, err
	}

	removeManagedFields(createdResource)
	return createdResource, nil
}

func (k8s *Kubectl) RemoveResourceDynamic(gvr schema.GroupVersionResource, isNamespaced bool, ns, name string) error {
	if gvr.Empty() {
		return fmt.Errorf("GroupVersionResource is empty")
	}
	if isNamespaced {
		return k8s.dynamicClient.Resource(gvr).Namespace(ns).Delete(context.TODO(), name, metav1.DeleteOptions{})
	} else {
		return k8s.dynamicClient.Resource(gvr).Delete(context.TODO(), name, metav1.DeleteOptions{})
	}
}

func (k8s *Kubectl) PatchResourceDynamic(gvr schema.GroupVersionResource, isNamespaced bool, ns, name string, patchType types.PatchType, patchData []byte) (*unstructured.Unstructured, error) {
	if gvr.Empty() {
		return nil, fmt.Errorf("GroupVersionResource is empty")
	}
	var obj *unstructured.Unstructured
	var err error
	if isNamespaced {
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
func (k8s *Kubectl) UpdateResourceDynamic(gvr schema.GroupVersionResource, isNamespaced bool, resource *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	if gvr.Empty() {
		return nil, fmt.Errorf("GroupVersionResource is empty")
	}
	var updatedResource *unstructured.Unstructured
	var err error
	if isNamespaced {
		updatedResource, err = k8s.dynamicClient.Resource(gvr).Namespace(resource.GetNamespace()).Update(context.TODO(), resource, metav1.UpdateOptions{})

	} else {
		updatedResource, err = k8s.dynamicClient.Resource(gvr).Update(context.TODO(), resource, metav1.UpdateOptions{})
	}

	if err != nil {
		return nil, fmt.Errorf("无法更新资源: %v", err)
	}
	removeManagedFields(updatedResource)
	return updatedResource, nil
}

// GetGVR 返回对应 string 的 GroupVersionResource
func (k8s *Kubectl) GetGVR(kind string) (gvr schema.GroupVersionResource, namespaced bool) {
	for _, resource := range k8s.apiResources {
		if resource.Kind == kind {
			version := resource.Version
			gvr = schema.GroupVersionResource{
				Group:    resource.Group,
				Version:  version,
				Resource: resource.Name, // 通常是 Kind 的复数形式
			}
			return gvr, resource.Namespaced
		}
	}
	return schema.GroupVersionResource{}, false
}
func (k8s *Kubectl) IsBuiltinResource(kind string) bool {
	for _, list := range k8s.apiResources {
		if list.Kind == kind {
			return true
		}
	}
	return false
}
