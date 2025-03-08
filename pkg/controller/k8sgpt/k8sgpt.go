package k8sgpt

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/k8sgpt/analysis"
	"k8s.io/klog/v2"
)

func RunAnalysis(c *gin.Context) {
	ctx := amis.GetContextWithUser(c)
	selectedCluster := amis.GetSelectedCluster(c)

	config, err := analysis.NewAnalysis(ctx,
		selectedCluster,
		// []string{"Pod", "Service", "Deployment", "ReplicaSet", "PersistentVolumeClaim", "Service", "Ingress", "StatefulSet", "CronJob", "Node", "ValidatingWebhookConfiguration", "MutatingWebhookConfiguration", "HorizontalPodAutoScaler", "PodDisruptionBudget", "NetworkPolicy", "Log"}, // Filter for these analyzers (e.g. Pod, PersistentVolumeClaim, Service, ReplicaSet)
		[]string{"Deployment"}, // Filter for these analyzers (e.g. Pod, PersistentVolumeClaim, Service, ReplicaSet)
		"*",
		"",
		true,
		1,
		true,
		true,
	)

	if err != nil {
		klog.Errorf("Error: %v", err)
	}
	defer config.Close()

	config.RunAnalysis()
	var output = "text"
	if err := config.GetAIResults(output, true); err != nil {
		color.Red("Error: %v", err)
		os.Exit(1)
	}
	// print results
	output_data, err := config.PrintOutput(output)
	if err != nil {
		color.Red("Error: %v", err)
	}
	statsData := config.PrintStats()
	fmt.Println(string(statsData))

	fmt.Println(string(output_data))

}
