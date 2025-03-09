package k8sgpt

import (
	"strings"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/k8sgpt/analysis"
	"github.com/weibaohui/k8m/pkg/k8sgpt/kubernetes"
	"github.com/weibaohui/kom/kom"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
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

type AnalysisConfig struct {
	Filters        []string
	Explain        bool
	Async          bool
	Namespace      string
	labelSelector  string
	maxConcurrency int
	withDoc        bool
	withStats      bool
}

func createAnalysisConfig(c *gin.Context) *AnalysisConfig {
	kind := c.Param("kind")
	if kind != "" {
		return &AnalysisConfig{
			Filters:        []string{kind},
			Explain:        c.Param("explain") == "true",
			Async:          false,
			Namespace:      "*",
			labelSelector:  "",
			maxConcurrency: 1,
			withDoc:        true,
			withStats:      false,
		}
	}
	return &AnalysisConfig{
		Filters: []string{"Pod", "Service", "Deployment", "ReplicaSet", "PersistentVolumeClaim",
			"Ingress", "StatefulSet", "CronJob", "Node", "ValidatingWebhookConfiguration",
			"MutatingWebhookConfiguration", "HorizontalPodAutoScaler", "PodDisruptionBudget", "NetworkPolicy"},
		Explain:        true,
		Async:          strings.Contains(c.Request.URL.Path, "async"),
		Namespace:      "*",
		labelSelector:  "",
		maxConcurrency: 1,
		withDoc:        true,
		withStats:      false,
	}
}

func writeAnalysisResult(c *gin.Context, config *analysis.Analysis) {
	output := "markdown"
	outputData, err := config.PrintOutput(output)
	if err != nil {
		color.Red("Error: %v", err)
		return
	}
	amis.WriteJsonData(c, gin.H{"result": string(outputData)})
}

func RunAnalysis(c *gin.Context) {
	ctx := amis.GetContextWithUser(c)
	cfg := createAnalysisConfig(c)

	config, err := analysis.NewAnalysis(
		ctx,
		amis.GetSelectedCluster(c),
		cfg.Filters,
		cfg.Namespace,
		cfg.labelSelector,
		cfg.Explain,
		cfg.maxConcurrency,
		cfg.withDoc,
		cfg.withStats,
	)

	if err != nil {
		klog.Errorf("Error: %v", err)
		return
	}
	defer config.Close()

	config.RunAnalysis()

	if cfg.Explain {
		if err := config.GetAIResults(true); err != nil {
			color.Red("Error: %v", err)
			return
		}
	}

	writeAnalysisResult(c, config)
}
func ClusterRunAnalysis(c *gin.Context) {
	handleAnalysisRequest(c, true)
}

func AsyncClusterRunAnalysis(c *gin.Context) {
	handleAnalysisRequest(c, false)
}

func handleAnalysisRequest(c *gin.Context, explain bool) {
	ctx := amis.GetContextWithUser(c)
	cfg := createAnalysisConfig(c)
	cfg.Explain = explain

	config, err := analysis.NewAnalysis(
		ctx,
		amis.GetSelectedCluster(c),
		cfg.Filters,
		cfg.Namespace,
		cfg.labelSelector,
		cfg.Explain,
		cfg.maxConcurrency,
		cfg.withDoc,
		cfg.withStats,
	)

	if err != nil {
		klog.Errorf("Error: %v", err)
		return
	}
	defer config.Close()

	config.RunAnalysis()

	if cfg.Explain {
		if err := config.GetAIResults(true); err != nil {
			color.Red("Error: %v", err)
			return
		}
	}

	writeAnalysisResult(c, config)
}
