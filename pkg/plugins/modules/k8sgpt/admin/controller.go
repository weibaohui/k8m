package admin

import (
	"fmt"

	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/k8sgpt/analysis"
	"github.com/weibaohui/k8m/pkg/k8sgpt/kubernetes"
	"github.com/weibaohui/k8m/pkg/response"
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/kom"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Controller struct{}

func createAnalysisConfig(c *response.Context) *analysis.Analysis {
	ctx := amis.GetContextWithUser(c)
	clusterID := ""
	clusterIDBase64 := c.Param("cluster")
	if clusterIDBase64 != "" {
		if id, err := utils.DecodeBase64(clusterIDBase64); err == nil {
			if id != "" {
				clusterID = id
			}
		}
	}

	ns := c.Param("ns")
	if ns == "" {
		ns = "*"
	}
	cfg := &analysis.Analysis{
		ClusterID:      clusterID,
		Context:        ctx,
		Namespace:      ns,
		LabelSelector:  "",
		MaxConcurrency: 1,
		WithDoc:        true,
		WithStats:      false,
	}

	return cfg
}

func (cc *Controller) GetFields(c *response.Context) {
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

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

func (cc *Controller) ResourceRunAnalysis(c *response.Context) {
	cfg := createAnalysisConfig(c)
	kind := c.Param("kind")
	cfg.Filters = []string{kind}
	result, err := analysis.Run(cfg)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, result)
}

func (cc *Controller) ClusterRunAnalysis(c *response.Context) {
	userCluster := c.Param("user_cluster")
	if userCluster != "" {
		if id, err := utils.UrlSafeBase64Decode(userCluster); err == nil {
			if string(id) != "" {
				userCluster = string(id)
			}
		}
	}
	cfg := createAnalysisConfig(c)
	cfg.ClusterID = userCluster
	cfg.Context = utils.GetContextWithAdmin()
	if !service.ClusterService().IsConnected(cfg.ClusterID) {
		amis.WriteJsonError(c, fmt.Errorf("集群 %s 未连接", cfg.ClusterID))
		return
	}
	go func() {
		cfg.Filters = []string{"Pod", "Service", "Deployment", "ReplicaSet", "PersistentVolumeClaim",
			"Ingress", "StatefulSet", "CronJob", "Node", "ValidatingWebhookConfiguration",
			"MutatingWebhookConfiguration", "HorizontalPodAutoScaler", "PodDisruptionBudget", "NetworkPolicy"}

		if result, err := analysis.Run(cfg); err == nil {
			service.ClusterService().SetClusterScanStatus(cfg.ClusterID, result)
		}
	}()

	amis.WriteJsonOKMsg(c, "后台执行，请稍后查看")
}

func (cc *Controller) GetClusterRunAnalysisResult(c *response.Context) {
	userCluster := c.Param("user_cluster")
	if userCluster != "" {
		if id, err := utils.UrlSafeBase64Decode(userCluster); err == nil {
			if string(id) != "" {
				userCluster = string(id)
			}
		}
	}
	cfg := createAnalysisConfig(c)
	cfg.ClusterID = userCluster
	scanResult := service.ClusterService().GetClusterScanResult(cfg.ClusterID)
	if scanResult == nil {
		amis.WriteJsonOKMsg(c, "暂无数据，请先点击执行检查")
		return
	}
	amis.WriteJsonData(c, scanResult)
}
