package kubernetes

import (
	openapi_v2 "github.com/google/gnostic/openapiv2"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type K8sApiReference struct {
	ApiVersion    schema.GroupVersion
	Kind          string
	OpenapiSchema *openapi_v2.Document
}
