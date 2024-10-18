package kubectl

import (
	"github.com/weibaohui/k8m/internal/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
	Error        error
	RowsAffected int64
	Statement    *Statement
	Namespace    string
	Name         string
	Group        string
	Version      string
	Kind         string
	GVR          schema.GroupVersionResource
	Namespaced   bool
	ListOptions  *metav1.ListOptions
	Type         StatementType // list get create update remove
	Resource     string
}

func (s *Statement) SetNamespace(ns string) *Statement {
	s.Namespace = ns
	return s
}
func (s *Statement) SetName(name string) *Statement {
	s.Name = name
	return s
}
func (s *Statement) SetGroup(group string) *Statement {
	s.Group = group
	return s
}
func (s *Statement) SetVersion(version string) *Statement {
	s.Version = version
	return s
}
func (s *Statement) SetKind(kind string) *Statement {
	s.Kind = kind
	return s
}
func (s *Statement) SetType(t StatementType) *Statement {
	s.Type = t
	return s
}
func (s *Statement) SetListOptions(opts *metav1.ListOptions) *Statement {
	s.ListOptions = opts
	return s
}
func (s *Statement) SetGVR(gvr schema.GroupVersionResource) *Statement {
	s.GVR = gvr
	s.Group = gvr.Group
	s.Version = gvr.Version
	s.Resource = gvr.Resource

	return s
}
func (s *Statement) SetNamespaced(b bool) *Statement {
	s.Namespaced = b
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
func (s *Statement) SetStatement(stmt *Statement) *Statement {
	s.Statement = stmt
	return s
}
func (s *Statement) SetResource(resource string) *Statement {
	s.Resource = resource
	return s
}

func (s *Statement) String() string {
	return utils.ToJSON(s)
}
