package k8sgpt

import (
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
func RunAnalysis(c *gin.Context) {
	ctx := amis.GetContextWithUser(c)
	kind := c.Param("kind")
	selectedCluster := amis.GetSelectedCluster(c)

	config, err := analysis.NewAnalysis(ctx,
		selectedCluster,
		// []string{"Pod", "Service", "Deployment", "ReplicaSet", "PersistentVolumeClaim", "Service", "Ingress", "StatefulSet", "CronJob", "Node", "ValidatingWebhookConfiguration", "MutatingWebhookConfiguration", "HorizontalPodAutoScaler", "PodDisruptionBudget", "NetworkPolicy", "Log"}, // Filter for these analyzers (e.g. Pod, PersistentVolumeClaim, Service, ReplicaSet)
		[]string{kind}, // Filter for these analyzers (e.g. Pod, PersistentVolumeClaim, Service, ReplicaSet)
		"*",
		"",
		true,
		1,
		true,
		false,
	)

	if err != nil {
		klog.Errorf("Error: %v", err)
		return
	}
	defer config.Close()

	config.RunAnalysis()
	var output = "markdown"
	if err := config.GetAIResults(true); err != nil {
		color.Red("Error: %v", err)
		return

	}
	// print results
	output_data, err := config.PrintOutput(output)
	if err != nil {
		color.Red("Error: %v", err)
		return

	}
	// statsData := config.PrintStats()
	// fmt.Println(string(statsData))

	amis.WriteJsonData(c, gin.H{
		"result": string(output_data),
	})

}
func ClusterRunAnalysis(c *gin.Context) {
	ctx := amis.GetContextWithUser(c)
	selectedCluster := amis.GetSelectedCluster(c)

	config, err := analysis.NewAnalysis(ctx,
		selectedCluster,
		[]string{"Pod", "Service", "Deployment", "ReplicaSet", "PersistentVolumeClaim", "Service", "Ingress", "StatefulSet", "CronJob", "Node", "ValidatingWebhookConfiguration", "MutatingWebhookConfiguration", "HorizontalPodAutoScaler", "PodDisruptionBudget", "NetworkPolicy"}, // Filter for these analyzers (e.g. Pod, PersistentVolumeClaim, Service, ReplicaSet)
		"*",
		"",
		true,
		1,
		true,
		false,
	)

	if err != nil {
		klog.Errorf("Error: %v", err)
		return
	}
	defer config.Close()

	config.RunAnalysis()

	var output = "markdown"
	if err := config.GetAIResults(true); err != nil {
		color.Red("Error: %v", err)
		return

	}
	// print results
	output_data, err := config.PrintOutput(output)
	if err != nil {
		color.Red("Error: %v", err)
		return

	}
	// statsData := config.PrintStats()
	// fmt.Println(string(statsData))

	amis.WriteJsonData(c, gin.H{
		"result": string(output_data),
	})

}
func AsyncClusterRunAnalysis(c *gin.Context) {
	ctx := amis.GetContextWithUser(c)
	selectedCluster := amis.GetSelectedCluster(c)

	config, err := analysis.NewAnalysis(ctx,
		selectedCluster,
		[]string{"Pod", "Service", "Deployment", "ReplicaSet", "PersistentVolumeClaim", "Service", "Ingress", "StatefulSet", "CronJob", "Node", "ValidatingWebhookConfiguration", "MutatingWebhookConfiguration", "HorizontalPodAutoScaler", "PodDisruptionBudget", "NetworkPolicy"}, // Filter for these analyzers (e.g. Pod, PersistentVolumeClaim, Service, ReplicaSet)
		"*",
		"",
		false,
		1,
		true,
		false,
	)

	if err != nil {
		klog.Errorf("Error: %v", err)
		return
	}
	defer config.Close()

	config.RunAnalysis()

	// 完成巡检结果存放在config.Results
	// 改为异步获取AI解答
	//

	var output = "markdown"
	if err := config.GetAIResults(true); err != nil {
		color.Red("Error: %v", err)
		return

	}
	// print results
	output_data, err := config.PrintOutput(output)
	if err != nil {
		color.Red("Error: %v", err)
		return

	}
	// statsData := config.PrintStats()
	// fmt.Println(string(statsData))

	amis.WriteJsonData(c, gin.H{
		"result": string(output_data),
	})

}
