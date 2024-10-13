package kubectl

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (k8s *Kubectl) FetchCRD(crd *unstructured.Unstructured, ns, name string) (*unstructured.Unstructured, error) {
	gvr := k8s.getGRVFromCRD(crd)
	// 检查CRD是否是Namespaced
	isNamespaced := crd.Object["spec"].(map[string]interface{})["scope"].(string) == "Namespaced"

	if ns == "" && isNamespaced {
		ns = "default" // 默认命名空间
	}
	return k8s.GetResourceDynamic(gvr, isNamespaced, ns, name)
}
func (k8s *Kubectl) RemoveCRD(crd *unstructured.Unstructured, ns, name string) error {
	gvr := k8s.getGRVFromCRD(crd)
	// 检查CRD是否是Namespaced
	isNamespaced := crd.Object["spec"].(map[string]interface{})["scope"].(string) == "Namespaced"

	if ns == "" && isNamespaced {
		ns = "default" // 默认命名空间
	}
	return k8s.RemoveResourceDynamic(gvr, isNamespaced, ns, name)
}
func (k8s *Kubectl) UpdateCRD(crd *unstructured.Unstructured, res *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	gvr := k8s.getGRVFromCRD(crd)
	// 检查CRD是否是Namespaced
	isNamespaced := crd.Object["spec"].(map[string]interface{})["scope"].(string) == "Namespaced"

	if res.GetNamespace() == "" && isNamespaced {
		res.SetNamespace("default") // 默认命名空间
	}
	return k8s.UpdateResourceDynamic(gvr, isNamespaced, res)
}
func (k8s *Kubectl) ListCRD(crd *unstructured.Unstructured, ns string) ([]unstructured.Unstructured, error) {
	gvr := k8s.getGRVFromCRD(crd)
	// 检查CRD是否是Namespaced
	isNamespaced := crd.Object["spec"].(map[string]interface{})["scope"].(string) == "Namespaced"

	if ns == "" && isNamespaced {
		ns = "default" // 默认命名空间
	}
	return k8s.ListResourcesDynamic(gvr, isNamespaced, ns)
}

func (k8s *Kubectl) GetCRD(kind string, group string) (*unstructured.Unstructured, error) {
	crdList, err := k8s.ListResources("CustomResourceDefinition", "")
	if err != nil {
		return nil, err
	}
	for _, crd := range crdList {
		spec, found, err := unstructured.NestedMap(crd.Object, "spec")
		if err != nil || !found {
			continue
		}
		crdKind, found, err := unstructured.NestedString(spec, "names", "kind")
		if err != nil || !found {
			continue
		}
		crdGroup, found, err := unstructured.NestedString(spec, "group")
		if err != nil || !found {
			continue
		}
		if crdKind != kind || crdGroup != group {
			continue
		}
		return &crd, nil
	}
	return nil, fmt.Errorf("crd %s.%s not found", kind, group)
}

func (k8s *Kubectl) getGRVFromCRD(crd *unstructured.Unstructured) schema.GroupVersionResource {
	// 提取 GVR
	group := crd.Object["spec"].(map[string]interface{})["group"].(string)
	version := crd.Object["spec"].(map[string]interface{})["versions"].([]interface{})[0].(map[string]interface{})["name"].(string)
	resource := crd.Object["spec"].(map[string]interface{})["names"].(map[string]interface{})["plural"].(string)

	gvr := schema.GroupVersionResource{
		Group:    group,
		Version:  version,
		Resource: resource,
	}
	return gvr
}
func (k8s *Kubectl) DeleteCRD(crd *unstructured.Unstructured, obj *unstructured.Unstructured) (result string) {
	gvr := k8s.getGRVFromCRD(crd)
	// 检查CRD是否是Namespaced
	isNamespaced := crd.Object["spec"].(map[string]interface{})["scope"].(string) == "Namespaced"
	ns := obj.GetNamespace()
	name := obj.GetName()

	if ns == "" && isNamespaced {
		ns = "default" // 默认命名空间
		obj.SetNamespace(ns)
	}
	err := k8s.RemoveResourceDynamic(gvr, isNamespaced, ns, name)
	if err != nil {
		result = fmt.Sprintf("%s/%s deleted error:%v", obj.GetKind(), obj.GetName(), err)
	} else {
		result = fmt.Sprintf("%s/%s deleted", obj.GetKind(), obj.GetName())
	}
	return result
}
func (k8s *Kubectl) ApplyCRD(crd *unstructured.Unstructured, obj *unstructured.Unstructured) (result string) {
	gvr := k8s.getGRVFromCRD(crd)
	// 检查CRD是否是Namespaced
	isNamespaced := crd.Object["spec"].(map[string]interface{})["scope"].(string) == "Namespaced"
	ns := obj.GetNamespace()
	name := obj.GetName()
	kind := obj.GetKind()

	if ns == "" && isNamespaced {
		ns = "default" // 默认命名空间
		obj.SetNamespace(ns)
	}
	exist, err := k8s.GetResourceDynamic(gvr, isNamespaced, ns, name)
	if err == nil && exist != nil && exist.GetName() != "" {
		// 已经存在资源，那么就更新
		obj.SetResourceVersion(exist.GetResourceVersion())
		_, err := k8s.UpdateResourceDynamic(gvr, isNamespaced, obj)
		if err != nil {
			result = fmt.Sprintf("更新CRD应用失败：%v", err.Error())
			return result
		}
		result = fmt.Sprintf("%s/%s updated", kind, name)
	} else {
		// 不存在，那么就创建
		_, err := k8s.CreateResourceDynamic(gvr, isNamespaced, obj)
		if err != nil {
			result = fmt.Sprintf("创建CRD应用失败：%v %s/%s %v", err.Error(), gvr.GroupResource(), name, isNamespaced)
			return result
		}
		result = fmt.Sprintf("%s/%s created", kind, name)
	}
	return result
}
