package k8sgpt

import (
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/k8sgpt/analysis"
	"github.com/weibaohui/k8m/pkg/k8sgpt/kubernetes"
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/kom"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func GetFields(c *gin.Context) {
	selectedCluster := amis.GetSelectedCluster(c)

	kind := "Deployment"
	apiDoc := kubernetes.K8sApiReference{
		Kind: kind,
		ApiVersion: schema.GroupVersion{
			Group:   "apps",
			Version: "v1",
		},
		OpenapiSchema: kom.Cluster(selectedCluster).Status().OpenAPISchema(),
	}

	v2 := apiDoc.GetApiDocV2("spec.template.spec.containers.imagePullPolicy")
	amis.WriteJsonData(c, v2)
}

func createAnalysisConfig(c *gin.Context) *analysis.Analysis {
	ctx := amis.GetContextWithUser(c)
	clusterID := amis.GetSelectedCluster(c)
	cfg := &analysis.Analysis{
		ClusterID:      clusterID,
		Context:        ctx,
		Namespace:      "*",
		LabelSelector:  "",
		MaxConcurrency: 1,
		WithDoc:        true,
		WithStats:      false,
	}
	if client, err := service.AIService().DefaultClient(); err == nil {
		cfg.AIClient = client
	}
	return cfg
}

func ResourceRunAnalysis(c *gin.Context) {
	cfg := createAnalysisConfig(c)
	kind := c.Param("kind")
	cfg.Filters = []string{kind}
	result, err := analysis.Run(cfg)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, gin.H{"result": string(result)})
}
func ClusterRunAnalysis(c *gin.Context) {
	cfg := createAnalysisConfig(c)
	cfg.Filters = []string{"Pod", "Service", "Deployment", "ReplicaSet", "PersistentVolumeClaim",
		"Ingress", "StatefulSet", "CronJob", "Node", "ValidatingWebhookConfiguration",
		"MutatingWebhookConfiguration", "HorizontalPodAutoScaler", "PodDisruptionBudget", "NetworkPolicy"}
	cfg.Explain = false
	result, err := analysis.Run(cfg)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, gin.H{"result": string(result)})
}

func AsyncClusterRunAnalysis(c *gin.Context) {
	cfg := createAnalysisConfig(c)
	cfg.Filters = []string{"Pod", "Service", "Deployment", "ReplicaSet", "PersistentVolumeClaim",
		"Ingress", "StatefulSet", "CronJob", "Node", "ValidatingWebhookConfiguration",
		"MutatingWebhookConfiguration", "HorizontalPodAutoScaler", "PodDisruptionBudget", "NetworkPolicy"}
	cfg.Explain = true

	result, err := analysis.Run(cfg)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, gin.H{"result": string(result)})
}
