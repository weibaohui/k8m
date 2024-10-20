package kubectl

import (
	"context"

	"github.com/weibaohui/k8m/pkg/comm/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
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
	Group         string
	Version       string
	Kind          string
	GVR           schema.GroupVersionResource
	Namespaced    bool
	ListOptions   *metav1.ListOptions
	Type          StatementType // list get create update remove
	Resource      string
	Context       context.Context
	client        *kubernetes.Clientset
	config        *rest.Config
	DynamicClient dynamic.Interface
	Dest          interface{}
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
	s.Group = gvr.Group
	s.Version = gvr.Version
	s.Resource = gvr.Resource

	return s
}

func (s *Statement) SetError(err error) *Statement {
	s.Error = err
	return s
}
func (s *Statement) SetRowsAffected(rows int64) *Statement {
	s.RowsAffected = rows
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
		Group:       s.Group,
		Version:     s.Version,
		Kind:        s.Kind,
		GVR:         s.GVR,
		Resource:    s.Resource,
		Namespaced:  s.Namespaced,
		ListOptions: s.ListOptions,
		Type:        s.Type,
		Context:     s.Context,
	}

	return newStmt
}

func (s *Statement) SetGVKs(gvks []schema.GroupVersionKind, version ...string) *Statement {
	var v string
	if len(version) > 0 {
		// 指定了版本
		v = version[0]
		for _, gvk := range gvks {
			if gvk.Version == v {
				s.Kind = gvk.Kind
				s.Group = gvk.Group
				s.Kind = gvk.Kind
				break
			}
		}
	} else {
		// 取第一个
		s.Kind = gvks[0].Kind
		s.Group = gvks[0].Group
		s.Version = gvks[0].Version
	}

	gvr, namespaced := s.GetGVR(s.Kind)
	s.setGVR(gvr)
	s.Namespaced = namespaced

	// 检查是否CRD，CRD需要检查scope
	if !s.IsBuiltinResource(s.Kind) {
		crd, err := s.GetCRD(context.TODO(), s.Kind, gvks[0].Group)
		if err != nil {
			return s
		}
		// 检查CRD是否是Namespaced
		s.Namespaced = crd.Object["spec"].(map[string]interface{})["scope"].(string) == "Namespaced"

	}

	return s
}
