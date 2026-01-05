package admin

import (
	"fmt"

	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/k8sgpt/analysis"
	"github.com/weibaohui/k8m/pkg/response"
	"github.com/weibaohui/k8m/pkg/service"
)

type Controller struct{}

func createAnalysisConfig(c *response.Context) *analysis.Analysis {
	ctx := amis.GetContextWithUser(c)
	clusterID := ""
	clusterIDBase64 := c.Param("cluster")
	if clusterIDBase64 != "" {
		if id, err := utils.UrlSafeBase64Decode(clusterIDBase64); err == nil {
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

func (cc *Controller) ClusterRunAnalysisMgm(c *response.Context) {
	clusterIDBase64 := c.Param("cluster")
	if clusterIDBase64 != "" {
		if id, err := utils.DecodeBase64(clusterIDBase64); err == nil {
			if id != "" {
				clusterIDBase64 = id
			}
		}
	}
	cfg := createAnalysisConfig(c)
	cfg.ClusterID = clusterIDBase64
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

func (cc *Controller) GetClusterRunAnalysisResultMgm(c *response.Context) {
	clusterIDBase64 := c.Param("cluster")
	if clusterIDBase64 != "" {
		if id, err := utils.DecodeBase64(clusterIDBase64); err == nil {
			if id != "" {
				clusterIDBase64 = id
			}
		}
	}
	cfg := createAnalysisConfig(c)
	cfg.ClusterID = clusterIDBase64
	scanResult := service.ClusterService().GetClusterScanResult(cfg.ClusterID)
	if scanResult == nil {
		amis.WriteJsonOKMsg(c, "暂无数据，请先点击执行检查")
		return
	}
	amis.WriteJsonData(c, scanResult)
}
