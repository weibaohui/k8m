package admin

import (
	"fmt"

	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/plugins/modules/k8sgpt/service/analysis"
	"github.com/weibaohui/k8m/pkg/response"
	"github.com/weibaohui/k8m/pkg/service"
)

type Controller struct{}

func createAnalysisConfig(c *response.Context) *analysis.Analysis {
	ctx := amis.GetContextWithUser(c)
	clusterID := ""
	clusterIdentifier := c.Param("cluster")
	if clusterIdentifier != "" {
		if id, err := service.ClusterService().ResolveClusterID(clusterIdentifier); err == nil && id != "" {
			clusterID = id
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
		if id, err := service.ClusterService().ResolveClusterID(userCluster); err == nil && id != "" {
			userCluster = id
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
			service.ClusterService().GetClusterByID(cfg.ClusterID).SetClusterScanStatus(result)
		}
	}()

	amis.WriteJsonOKMsg(c, "后台执行，请稍后查看")
}

func (cc *Controller) ClusterRunAnalysisMgm(c *response.Context) {
	clusterIdentifier := c.Param("cluster")
	clusterID, err := service.ClusterService().ResolveClusterID(clusterIdentifier)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	cfg := createAnalysisConfig(c)
	cfg.ClusterID = clusterID
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
			service.ClusterService().GetClusterByID(cfg.ClusterID).SetClusterScanStatus(result)
		}
	}()

	amis.WriteJsonOKMsg(c, "后台执行，请稍后查看")
}

func (cc *Controller) GetClusterRunAnalysisResultMgm(c *response.Context) {
	clusterIdentifier := c.Param("cluster")
	clusterID, err := service.ClusterService().ResolveClusterID(clusterIdentifier)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	cfg := createAnalysisConfig(c)
	cfg.ClusterID = clusterID
	scanResult := service.ClusterService().GetClusterByID(cfg.ClusterID).GetClusterScanResult()
	if scanResult == nil {
		amis.WriteJsonOKMsg(c, "暂无数据，请先点击执行检查")
		return
	}
	amis.WriteJsonData(c, scanResult)
}
