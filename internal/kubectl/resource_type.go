package kubectl

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ResourceType 定义 Kubernetes 资源的枚举类型
type ResourceType string

// 定义常用 Kubernetes 资源的枚举值
const (
	Pod                            ResourceType = "Pod"
	Service                        ResourceType = "Service"
	ReplicationController          ResourceType = "ReplicationController"
	Namespace                      ResourceType = "Namespace"
	Node                           ResourceType = "Node"
	PersistentVolume               ResourceType = "PersistentVolume"
	PersistentVolumeClaim          ResourceType = "PersistentVolumeClaim"
	ConfigMap                      ResourceType = "ConfigMap"
	Secret                         ResourceType = "Secret"
	ServiceAccount                 ResourceType = "ServiceAccount"
	Event                          ResourceType = "Event"
	Endpoints                      ResourceType = "Endpoints"
	LimitsRange                    ResourceType = "LimitRange"
	ResourceQuota                  ResourceType = "ResourceQuota"
	Deployment                     ResourceType = "Deployment"
	StatefulSet                    ResourceType = "StatefulSet"
	DaemonSet                      ResourceType = "DaemonSet"
	ReplicaSet                     ResourceType = "ReplicaSet"
	ControllerRevision             ResourceType = "ControllerRevision"
	HorizontalPodAutoscaler        ResourceType = "HorizontalPodAutoscaler"
	Job                            ResourceType = "Job"
	CronJob                        ResourceType = "CronJob"
	Ingress                        ResourceType = "Ingress"
	NetworkPolicy                  ResourceType = "NetworkPolicy"
	Role                           ResourceType = "Role"
	ClusterRole                    ResourceType = "ClusterRole"
	RoleBinding                    ResourceType = "RoleBinding"
	ClusterRoleBinding             ResourceType = "ClusterRoleBinding"
	StorageClass                   ResourceType = "StorageClass"
	VolumeAttachment               ResourceType = "VolumeAttachment"
	PodSecurityPolicy              ResourceType = "PodSecurityPolicy"
	ValidatingWebhookConfiguration ResourceType = "ValidatingWebhookConfiguration"
	MutatingWebhookConfiguration   ResourceType = "MutatingWebhookConfiguration"
	CustomResourceDefinition       ResourceType = "CustomResourceDefinition"
	Binding                        ResourceType = "Binding"
	ComponentStatus                ResourceType = "ComponentStatus"
	PodTemplate                    ResourceType = "PodTemplate"
	APIService                     ResourceType = "APIService"
	SelfSubjectReview              ResourceType = "SelfSubjectReview"
	TokenReview                    ResourceType = "TokenReview"
	LocalSubjectAccessReview       ResourceType = "LocalSubjectAccessReview"
	SelfSubjectAccessReview        ResourceType = "SelfSubjectAccessReview"
	SelfSubjectRulesReview         ResourceType = "SelfSubjectRulesReview"
	SubjectAccessReview            ResourceType = "SubjectAccessReview"
	CertificateSigningRequest      ResourceType = "CertificateSigningRequest"
	Lease                          ResourceType = "Lease"
	EndpointSlice                  ResourceType = "EndpointSlice"
	FlowSchema                     ResourceType = "FlowSchema"
	PriorityLevelConfiguration     ResourceType = "PriorityLevelConfiguration"
	IngressClass                   ResourceType = "IngressClass"
	RuntimeClass                   ResourceType = "RuntimeClass"
	PodDisruptionBudget            ResourceType = "PodDisruptionBudget"
	PriorityClass                  ResourceType = "PriorityClass"
	CSIDriver                      ResourceType = "CSIDriver"
	CSINode                        ResourceType = "CSINode"
	CSIStorageCapacity             ResourceType = "CSIStorageCapacity"
)

// resourceTypeMap 映射字符串到 ResourceType，包括单数、复数和别名
var resourceTypeMap = map[string]ResourceType{
	"pd":                             Pod,
	"pod":                            Pod,
	"pods":                           Pod,
	"svc":                            Service,
	"service":                        Service,
	"services":                       Service,
	"rc":                             ReplicationController,
	"replicationcontroller":          ReplicationController,
	"replicationcontrollers":         ReplicationController,
	"namespace":                      Namespace,
	"namespaces":                     Namespace,
	"ns":                             Namespace,
	"no":                             Node,
	"node":                           Node,
	"nodes":                          Node,
	"pv":                             PersistentVolume,
	"persistentvolume":               PersistentVolume,
	"persistentvolumes":              PersistentVolume,
	"pvc":                            PersistentVolumeClaim,
	"persistentvolumeclaim":          PersistentVolumeClaim,
	"persistentvolumeclaims":         PersistentVolumeClaim,
	"cm":                             ConfigMap,
	"configmap":                      ConfigMap,
	"configmaps":                     ConfigMap,
	"secret":                         Secret,
	"secrets":                        Secret,
	"sa":                             ServiceAccount,
	"serviceaccount":                 ServiceAccount,
	"serviceaccounts":                ServiceAccount,
	"ev":                             Event,
	"event":                          Event,
	"events":                         Event,
	"ep":                             Endpoints,
	"endpoints":                      Endpoints,
	"limits":                         LimitsRange,
	"limitsrange":                    LimitsRange,
	"limitsranges":                   LimitsRange,
	"quota":                          ResourceQuota,
	"resourcequota":                  ResourceQuota,
	"resourcequotas":                 ResourceQuota,
	"deploy":                         Deployment,
	"deployment":                     Deployment,
	"deployments":                    Deployment,
	"sts":                            StatefulSet,
	"statefulset":                    StatefulSet,
	"statefulsets":                   StatefulSet,
	"ds":                             DaemonSet,
	"daemonset":                      DaemonSet,
	"daemonsets":                     DaemonSet,
	"rs":                             ReplicaSet,
	"replicaset":                     ReplicaSet,
	"replicasets":                    ReplicaSet,
	"controllerrevision":             ControllerRevision,
	"controllerrevisions":            ControllerRevision,
	"hpa":                            HorizontalPodAutoscaler,
	"horizontalpodautoscaler":        HorizontalPodAutoscaler,
	"horizontalpodautoscalers":       HorizontalPodAutoscaler,
	"job":                            Job,
	"jobs":                           Job,
	"cj":                             CronJob,
	"cronjob":                        CronJob,
	"cronjobs":                       CronJob,
	"ing":                            Ingress,
	"ingress":                        Ingress,
	"ingresses":                      Ingress,
	"netpol":                         NetworkPolicy,
	"networkpolicy":                  NetworkPolicy,
	"networkpolicies":                NetworkPolicy,
	"role":                           Role,
	"roles":                          Role,
	"clusterrole":                    ClusterRole,
	"clusterroles":                   ClusterRole,
	"rolebinding":                    RoleBinding,
	"rolebindings":                   RoleBinding,
	"clusterrolebinding":             ClusterRoleBinding,
	"clusterrolebindings":            ClusterRoleBinding,
	"sc":                             StorageClass,
	"storageclass":                   StorageClass,
	"storageclasses":                 StorageClass,
	"volumeattachment":               VolumeAttachment,
	"volumeattachments":              VolumeAttachment,
	"podsecuritypolicy":              PodSecurityPolicy,
	"podsecuritypolicies":            PodSecurityPolicy,
	"validatingwebhookconfiguration": ValidatingWebhookConfiguration,
	"mutatingwebhookconfiguration":   MutatingWebhookConfiguration,
	"crd":                            CustomResourceDefinition,
	"customresourcedefinition":       CustomResourceDefinition,
	"customresourcedefinitions":      CustomResourceDefinition,
	"binding":                        Binding,
	"bindings":                       Binding,
	"componentstatus":                ComponentStatus,
	"componentstatuses":              ComponentStatus,
	"cs":                             ComponentStatus,
	"podtemplate":                    PodTemplate,
	"podtemplates":                   PodTemplate,
	"apiservice":                     APIService,
	"apiservices":                    APIService,
	"selfsubjectreview":              SelfSubjectReview,
	"selfsubjectreviews":             SelfSubjectReview,
	"tokenreview":                    TokenReview,
	"tokenreviews":                   TokenReview,
	"localsubjectaccessreview":       LocalSubjectAccessReview,
	"localsubjectaccessreviews":      LocalSubjectAccessReview,
	"selfsubjectaccessreview":        SelfSubjectAccessReview,
	"selfsubjectaccessreviews":       SelfSubjectAccessReview,
	"selfsubjectrulesreview":         SelfSubjectRulesReview,
	"selfsubjectrulesreviews":        SelfSubjectRulesReview,
	"subjectaccessreview":            SubjectAccessReview,
	"subjectaccessreviews":           SubjectAccessReview,
	"certificatesigningrequest":      CertificateSigningRequest,
	"certificatesigningrequests":     CertificateSigningRequest,
	"csr":                            CertificateSigningRequest,
	"lease":                          Lease,
	"leases":                         Lease,
	"endpointslice":                  EndpointSlice,
	"endpointslices":                 EndpointSlice,
	"flowschema":                     FlowSchema,
	"flowschemas":                    FlowSchema,
	"prioritylevelconfiguration":     PriorityLevelConfiguration,
	"prioritylevelconfigurations":    PriorityLevelConfiguration,
	"ingressclass":                   IngressClass,
	"ingressclasses":                 IngressClass,
	"runtimeclass":                   RuntimeClass,
	"runtimeclasses":                 RuntimeClass,
	"pdb":                            PodDisruptionBudget,
	"poddisruptionbudget":            PodDisruptionBudget,
	"poddisruptionbudgets":           PodDisruptionBudget,
	"priorityclass":                  PriorityClass,
	"priorityclasses":                PriorityClass,
	"pc":                             PriorityClass,
	"csidriver":                      CSIDriver,
	"csidrivers":                     CSIDriver,
	"csinode":                        CSINode,
	"csinodes":                       CSINode,
	"csistoragecapacity":             CSIStorageCapacity,
	"csistoragecapacities":           CSIStorageCapacity,
}

// gvrMap 存储 ResourceType 到 GroupVersionResource 的映射
var gvrMap = map[ResourceType]schema.GroupVersionResource{
	Pod:                            {Group: "", Version: "v1", Resource: "pods"},
	Service:                        {Group: "", Version: "v1", Resource: "services"},
	ReplicationController:          {Group: "", Version: "v1", Resource: "replicationcontrollers"},
	Namespace:                      {Group: "", Version: "v1", Resource: "namespaces"},
	Node:                           {Group: "", Version: "v1", Resource: "nodes"},
	PersistentVolume:               {Group: "", Version: "v1", Resource: "persistentvolumes"},
	PersistentVolumeClaim:          {Group: "", Version: "v1", Resource: "persistentvolumeclaims"},
	ConfigMap:                      {Group: "", Version: "v1", Resource: "configmaps"},
	Secret:                         {Group: "", Version: "v1", Resource: "secrets"},
	ServiceAccount:                 {Group: "", Version: "v1", Resource: "serviceaccounts"},
	Event:                          {Group: "", Version: "v1", Resource: "events"},
	Endpoints:                      {Group: "", Version: "v1", Resource: "endpoints"},
	LimitsRange:                    {Group: "", Version: "v1", Resource: "limitranges"},
	ResourceQuota:                  {Group: "", Version: "v1", Resource: "resourcequotas"},
	Deployment:                     {Group: "apps", Version: "v1", Resource: "deployments"},
	StatefulSet:                    {Group: "apps", Version: "v1", Resource: "statefulsets"},
	DaemonSet:                      {Group: "apps", Version: "v1", Resource: "daemonsets"},
	ReplicaSet:                     {Group: "apps", Version: "v1", Resource: "replicasets"},
	ControllerRevision:             {Group: "apps", Version: "v1", Resource: "controllerrevisions"},
	HorizontalPodAutoscaler:        {Group: "autoscaling", Version: "v1", Resource: "horizontalpodautoscalers"},
	Job:                            {Group: "batch", Version: "v1", Resource: "jobs"},
	CronJob:                        {Group: "batch", Version: "v1beta1", Resource: "cronjobs"},
	Ingress:                        {Group: "networking.k8s.io", Version: "v1", Resource: "ingresses"},
	NetworkPolicy:                  {Group: "networking.k8s.io", Version: "v1", Resource: "networkpolicies"},
	Role:                           {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "roles"},
	ClusterRole:                    {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterroles"},
	RoleBinding:                    {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebindings"},
	ClusterRoleBinding:             {Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebindings"},
	StorageClass:                   {Group: "storage.k8s.io", Version: "v1", Resource: "storageclasses"},
	VolumeAttachment:               {Group: "storage.k8s.io", Version: "v1", Resource: "volumeattachments"},
	PodSecurityPolicy:              {Group: "policy", Version: "v1beta1", Resource: "podsecuritypolicies"},
	ValidatingWebhookConfiguration: {Group: "admissionregistration.k8s.io", Version: "v1", Resource: "validatingwebhookconfigurations"},
	MutatingWebhookConfiguration:   {Group: "admissionregistration.k8s.io", Version: "v1", Resource: "mutatingwebhookconfigurations"},
	CustomResourceDefinition:       {Group: "apiextensions.k8s.io", Version: "v1", Resource: "customresourcedefinitions"},
	Binding:                        {Group: "", Version: "v1", Resource: "bindings"},
	ComponentStatus:                {Group: "", Version: "v1", Resource: "componentstatuses"},
	PodTemplate:                    {Group: "", Version: "v1", Resource: "podtemplates"},
	APIService:                     {Group: "apiregistration.k8s.io", Version: "v1", Resource: "apiservices"},
	SelfSubjectReview:              {Group: "authentication.k8s.io", Version: "v1", Resource: "selfsubjectreviews"},
	TokenReview:                    {Group: "authentication.k8s.io", Version: "v1", Resource: "tokenreviews"},
	LocalSubjectAccessReview:       {Group: "authorization.k8s.io", Version: "v1", Resource: "localsubjectaccessreviews"},
	SelfSubjectAccessReview:        {Group: "authorization.k8s.io", Version: "v1", Resource: "selfsubjectaccessreviews"},
	SelfSubjectRulesReview:         {Group: "authorization.k8s.io", Version: "v1", Resource: "selfsubjectrulesreviews"},
	SubjectAccessReview:            {Group: "authorization.k8s.io", Version: "v1", Resource: "subjectaccessreviews"},
	CertificateSigningRequest:      {Group: "certificates.k8s.io", Version: "v1", Resource: "certificatesigningrequests"},
	Lease:                          {Group: "coordination.k8s.io", Version: "v1", Resource: "leases"},
	EndpointSlice:                  {Group: "discovery.k8s.io", Version: "v1", Resource: "endpointslices"},
	FlowSchema:                     {Group: "flowcontrol.apiserver.k8s.io", Version: "v1beta3", Resource: "flowschemas"},
	PriorityLevelConfiguration:     {Group: "flowcontrol.apiserver.k8s.io", Version: "v1beta3", Resource: "prioritylevelconfigurations"},
	IngressClass:                   {Group: "networking.k8s.io", Version: "v1", Resource: "ingressclasses"},
	RuntimeClass:                   {Group: "node.k8s.io", Version: "v1", Resource: "runtimeclasses"},
	PodDisruptionBudget:            {Group: "policy", Version: "v1", Resource: "poddisruptionbudgets"},
	PriorityClass:                  {Group: "scheduling.k8s.io", Version: "v1", Resource: "priorityclasses"},
	CSIDriver:                      {Group: "storage.k8s.io", Version: "v1", Resource: "csidrivers"},
	CSINode:                        {Group: "storage.k8s.io", Version: "v1", Resource: "csinodes"},
	CSIStorageCapacity:             {Group: "storage.k8s.io", Version: "v1", Resource: "csistoragecapacities"},
}

// 判断相关资源是否为Namespace级的
var nsMap = map[ResourceType]bool{
	Pod:                            true,
	Service:                        true,
	ReplicationController:          true,
	Namespace:                      false,
	Node:                           false,
	PersistentVolume:               false,
	PersistentVolumeClaim:          true,
	ConfigMap:                      true,
	Secret:                         true,
	ServiceAccount:                 true,
	Event:                          true,
	Endpoints:                      true,
	LimitsRange:                    true,
	ResourceQuota:                  true,
	Deployment:                     true,
	StatefulSet:                    true,
	DaemonSet:                      true,
	ReplicaSet:                     true,
	ControllerRevision:             true,
	HorizontalPodAutoscaler:        true,
	Job:                            true,
	CronJob:                        true,
	Ingress:                        true,
	NetworkPolicy:                  true,
	Role:                           true,
	ClusterRole:                    false,
	RoleBinding:                    true,
	ClusterRoleBinding:             false,
	StorageClass:                   false,
	VolumeAttachment:               false,
	PodSecurityPolicy:              false,
	ValidatingWebhookConfiguration: false,
	MutatingWebhookConfiguration:   false,
	CustomResourceDefinition:       false,
	Binding:                        true,
	ComponentStatus:                false,
	PodTemplate:                    true,
	APIService:                     false,
	SelfSubjectReview:              false,
	TokenReview:                    false,
	LocalSubjectAccessReview:       false,
	SelfSubjectAccessReview:        false,
	SelfSubjectRulesReview:         false,
	SubjectAccessReview:            false,
	CertificateSigningRequest:      false,
	Lease:                          true,
	EndpointSlice:                  true,
	FlowSchema:                     false,
	PriorityLevelConfiguration:     false,
	IngressClass:                   false,
	RuntimeClass:                   true,
	PodDisruptionBudget:            true,
	PriorityClass:                  false,
	CSIDriver:                      false,
	CSINode:                        false,
	CSIStorageCapacity:             false,
}

// GetGVR 返回对应 ResourceType 的 GroupVersionResource
func (rt ResourceType) GetGVR() schema.GroupVersionResource {
	if gvr, exists := gvrMap[rt]; exists {
		return gvr
	}
	return schema.GroupVersionResource{}
}

func (rt ResourceType) IsNamespaced() bool {
	if nsd, exists := nsMap[rt]; exists {
		return nsd
	}
	return false
}
func ParseResourceType(rt string) (ResourceType, error) {
	rt = strings.ToLower(rt)
	if gvr, exists := resourceTypeMap[rt]; exists {
		return gvr, nil
	}
	return "", fmt.Errorf("unknown resource type: %s", rt)
}
