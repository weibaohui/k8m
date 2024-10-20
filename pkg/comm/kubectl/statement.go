package kubectl

import (
	"context"
	"fmt"

	"github.com/weibaohui/k8m/pkg/comm/utils"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
)

type StatementType string

const (
	Query  StatementType = "query"
	Update StatementType = "update"
	Patch  StatementType = "patch"
	Delete StatementType = "delete"
	Create StatementType = "create"
)

type Statement struct {
	*Kubectl
	Error         error
	RowsAffected  int64
	Statement     *Statement
	Namespace     string
	Name          string
	GVR           schema.GroupVersionResource
	GVK           schema.GroupVersionKind
	Namespaced    bool
	ListOptions   *metav1.ListOptions
	Type          StatementType // list get create update remove
	Context       context.Context
	client        *kubernetes.Clientset
	config        *rest.Config
	DynamicClient dynamic.Interface
	Dest          interface{}
	Unstructured  bool // 返回Unstructured，适用于没有定义结构体的CRD
}

func (s *Statement) SetNamespace(ns string) *Statement {
	s.Namespace = ns
	return s
}
func (s *Statement) SetName(name string) *Statement {
	s.Name = name
	return s
}

func (s *Statement) SetType(t StatementType) *Statement {
	s.Type = t
	return s
}

func (s *Statement) setGVR(gvr schema.GroupVersionResource) *Statement {
	s.GVR = gvr
	return s
}

func (s *Statement) SetDest(dest interface{}) *Statement {
	s.Dest = dest
	return s
}

func (s *Statement) String() string {
	return utils.ToJSON(s)
}
func (s *Statement) clone() *Statement {
	newStmt := &Statement{
		Namespace:   s.Namespace,
		Name:        s.Name,
		GVR:         s.GVR,
		GVK:         s.GVK,
		Namespaced:  s.Namespaced,
		ListOptions: s.ListOptions,
		Type:        s.Type,
		Context:     s.Context,
	}

	return newStmt
}

func (s *Statement) ParseGVKs(gvks []schema.GroupVersionKind, versions ...string) *Statement {

	// 获取单个GVK
	gvk := s.GetParsedGVK(gvks, versions...)
	s.GVK = gvk

	// 获取GVR
	if s.IsBuiltinResource(gvk.Kind) {
		// 内置资源
		s.GVR, s.Namespaced = s.GetGVR(gvk.Kind)

	} else {
		crd, err := s.GetCRD(s.Context, gvk.Kind, gvk.Group)
		if err != nil {
			s.Error = err
			return s
		}
		// 检查CRD是否是Namespaced
		s.Namespaced = crd.Object["spec"].(map[string]interface{})["scope"].(string) == "Namespaced"
		s.GVR = s.getGRVFromCRD(crd)

	}

	return s
}
func (s *Statement) GetParsedGVK(gvks []schema.GroupVersionKind, versions ...string) (gvk schema.GroupVersionKind) {
	if len(gvks) == 0 {
		return schema.GroupVersionKind{}
	}
	if len(versions) > 0 {
		// 指定了版本
		v := versions[0]
		for _, g := range gvks {
			if g.Version == v {
				return schema.GroupVersionKind{
					Kind:    g.Kind,
					Group:   g.Group,
					Version: g.Version,
				}
			}
		}
	} else {
		// 取第一个
		return schema.GroupVersionKind{
			Kind:    gvks[0].Kind,
			Group:   gvks[0].Group,
			Version: gvks[0].Version,
		}

	}

	return
}

func (s *Statement) ParseNsNameFromRuntimeObj(obj runtime.Object) *Statement {
	// 获取元数据（比如Name和Namespace）
	accessor, err := meta.Accessor(obj)
	if err != nil {
		s.Error = err
		return s
	}
	s.Name = accessor.GetName()           // 获取资源的名称
	s.Namespace = accessor.GetNamespace() // 获取资源的命名空间
	return s
}

func (s *Statement) ParseGVKFromRuntimeObj(obj runtime.Object) *Statement {
	// 使用 scheme.Scheme.ObjectKinds() 获取 Kind
	gvks, _, err := scheme.Scheme.ObjectKinds(obj)
	if err != nil {
		s.Error = fmt.Errorf("error getting kind by scheme.Scheme.ObjectKinds : %v", err)
		return s
	}
	s.ParseGVKs(gvks)
	return s
}

func (s *Statement) ParseFromRuntimeObj(obj runtime.Object) *Statement {
	return s.
		ParseGVKFromRuntimeObj(obj).
		ParseNsNameFromRuntimeObj(obj)
}
