package kubectl

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

// ApplyYAML  解析并应用 YAML 字符串到 Kubernetes
func (k8s *Kubectl) ApplyYAML(yamlStr string) (result []string) {
	docs := splitYAML(yamlStr)

	for _, doc := range docs {
		if strings.TrimSpace(doc) == "" {
			continue
		}

		// 解析 YAML 到 Unstructured 对象
		var obj unstructured.Unstructured
		if err := yaml.Unmarshal([]byte(doc), &obj.Object); err != nil {
			result = append(result, fmt.Sprintf("YAML 解析失败: %v", err))
			continue
		}

		// 提取 Group, Version, Kind
		gvk := obj.GroupVersionKind()
		if gvk.Kind == "" || gvk.Version == "" {
			result = append(result, fmt.Sprintf("YAML 缺少必要的 Group, Version 或 Kind"))
			continue
		}

		builtin := k8s.IsBuiltinResource(gvk.Kind)
		if !builtin {
			crd, err := k8s.GetCRD(gvk.Kind, gvk.Group)
			if err != nil {
				result = append(result, fmt.Sprintf("%v", err))
			} else {
				crdResult := k8s.ApplyCRD(crd, &obj)
				result = append(result, crdResult)
			}
			continue

		}
		// 发现 string
		kind := gvk.Kind

		// 获取命名空间
		ns := obj.GetNamespace()
		if ns == "" && gvk.Kind != "Namespace" {
			ns = "default" // 默认命名空间
		}

		// 使用 CreateOrUpdateResource 应用资源
		_, err := k8s.CreateResource(kind, ns, &obj)
		if err != nil {
			if errors.IsAlreadyExists(err) {
				// 已经存在，更新
				existingResource, err := k8s.GetResource(kind, ns, obj.GetName())
				if err != nil {
					result = append(result, fmt.Sprintf("获取应用失败%v", err.Error()))
					continue
				}
				if existingResource != nil {
					obj.SetResourceVersion(existingResource.GetResourceVersion())
				}

				_, err = k8s.UpdateResource(kind, ns, &obj)
				if err != nil {
					result = append(result, fmt.Sprintf("更新应用失败：%v", err.Error()))
					continue
				}
				result = append(result, fmt.Sprintf("%s/%s updated", obj.GetKind(), obj.GetName()))

			} else {
				result = append(result, fmt.Sprintf("创建应用失败：%s/%s,%s,%v\n", obj.GetKind(), obj.GetName(), kind, err.Error()))
				continue
			}
		} else {
			result = append(result, fmt.Sprintf("%s/%s created", obj.GetKind(), obj.GetName()))
		}

	}

	return result
}
func (k8s *Kubectl) DeleteYAML(yamlStr string) (result []string) {
	docs := splitYAML(yamlStr)

	for _, doc := range docs {
		if strings.TrimSpace(doc) == "" {
			continue
		}

		// 解析 YAML 到 Unstructured 对象
		var obj unstructured.Unstructured
		if err := yaml.Unmarshal([]byte(doc), &obj.Object); err != nil {
			result = append(result, fmt.Sprintf("YAML 解析失败: %v", err))
			continue
		}

		// 提取 Group, Version, Kind
		gvk := obj.GroupVersionKind()
		if gvk.Kind == "" || gvk.Version == "" {
			result = append(result, fmt.Sprintf("YAML 缺少必要的 Group, Version 或 Kind"))
			continue
		}

		// 发现 string
		builtIn := k8s.IsBuiltinResource(gvk.Kind)
		if !builtIn {
			// CRD 类型资源
			crd, err := k8s.GetCRD(gvk.Kind, gvk.Group)
			if err != nil {
				result = append(result, fmt.Sprintf("%v", err))
			} else {
				// 确认为 CRD
				crdResult := k8s.DeleteCRD(crd, &obj)
				result = append(result, crdResult)
			}
			continue
		}

		// 获取命名空间
		ns := obj.GetNamespace()
		if ns == "" && gvk.Kind != "Namespace" {
			ns = "default" // 默认命名空间
		}

		err := k8s.DeleteResource(gvk.Kind, ns, obj.GetName())
		if err != nil {
			result = append(result, fmt.Sprintf("%s/%s deleted error:%v", obj.GetKind(), obj.GetName(), err))
		} else {
			result = append(result, fmt.Sprintf("%s/%s deleted", obj.GetKind(), obj.GetName()))
		}

	}

	return result
}
