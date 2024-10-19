package kubectl

import (
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/yaml"
)

func (k8s *Kubectl) ConvertUnstructuredToTypedObject(obj *unstructured.Unstructured, objType runtime.Object) error {
	decoder := scheme.Codecs.UniversalDeserializer()
	objBytes, err := obj.MarshalJSON()
	if err != nil {
		return fmt.Errorf("无法序列化 Unstructured 对象: %v", err)
	}

	_, _, err = decoder.Decode(objBytes, nil, obj)
	if err != nil {
		return fmt.Errorf("无法将 Unstructured 解码为具体类型: %v", err)
	}
	return nil
}

// ConvertToUnstructured 通用转换函数，将 runtime.Object 转换为 Unstructured
func (k8s *Kubectl) ConvertToUnstructured(obj interface{}) (*unstructured.Unstructured, error) {
	// 使用 DefaultUnstructuredConverter 转换为 map
	unstructuredMap, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
	if err != nil {
		return nil, err
	}

	// 获取资源类型和版本信息
	gvk, err := GetGVK(obj)
	if err != nil {
		return nil, err
	}

	// 创建 unstructured.Unstructured 对象并设置数据
	u := &unstructured.Unstructured{Object: unstructuredMap}
	u.SetGroupVersionKind(gvk)

	return u, nil
}

// GetGVK 获取对象的 GroupVersionKind
func GetGVK(obj interface{}) (schema.GroupVersionKind, error) {
	switch o := obj.(type) {
	case *unstructured.Unstructured:
		return o.GroupVersionKind(), nil
	case runtime.Object:
		return o.GetObjectKind().GroupVersionKind(), nil
	default:
		return schema.GroupVersionKind{}, fmt.Errorf("不支持的类型%v", o)
	}
}

// ConvertUnstructuredToYAML 将 Unstructured 对象转换为 YAML 字符串
func (k8s *Kubectl) ConvertUnstructuredToYAML(obj *unstructured.Unstructured) (string, error) {

	// Marshal Unstructured 对象为 JSON
	jsonBytes, err := obj.MarshalJSON()
	if err != nil {
		return "", fmt.Errorf("无法序列化 Unstructured 对象为 JSON: %v", err)
	}

	// 将 JSON 转换为 YAML
	yamlBytes, err := yaml.JSONToYAML(jsonBytes)
	if err != nil {
		return "", fmt.Errorf("无法将 JSON 转换为 YAML: %v", err)
	}

	return string(yamlBytes), nil
}
