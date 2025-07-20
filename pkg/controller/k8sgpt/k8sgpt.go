package k8sgpt

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/k8sgpt/analysis"
	"github.com/weibaohui/k8m/pkg/k8sgpt/kubernetes"
	"github.com/weibaohui/k8m/pkg/service"
	"github.com/weibaohui/kom/kom"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

type Controller struct{}

func RegisterRoutes(api *gin.RouterGroup) {
	ctrl := &Controller{}
	api.GET("/k8s_gpt/kind/:kind/run", ctrl.ResourceRunAnalysis)
	api.POST("/k8s_gpt/cluster/:user_cluster/run", ctrl.ClusterRunAnalysis)
	api.GET("/k8s_gpt/cluster/:user_cluster/result", ctrl.GetClusterRunAnalysisResult)
	api.GET("/k8s_gpt/var", ctrl.GetFields)
}

// @Summary 获取K8s资源字段信息
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/k8s_gpt/var [get]
func (cc *Controller) GetFields(c *gin.Context) {
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

func createAnalysisConfig(c *gin.Context) *analysis.Analysis {
	ctx := amis.GetContextWithUser(c)
	clusterID := ""
	clusterIDBase64 := c.Param("cluster") // 路径上传递的集群名称
	if clusterIDBase64 != "" {
		if id, err := utils.DecodeBase64(clusterIDBase64); err == nil {
			if id != "" {
				// 路径中传了集群名称
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

// @Summary 对指定资源类型运行K8sGPT分析
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param kind path string true "资源类型"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/k8s_gpt/kind/{kind}/run [get]
func (cc *Controller) ResourceRunAnalysis(c *gin.Context) {
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

// @Summary 对整个集群运行K8sGPT分析
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param user_cluster path string true "用户集群标识"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/k8s_gpt/cluster/{user_cluster}/run [post]
func (cc *Controller) ClusterRunAnalysis(c *gin.Context) {
	userCluster := c.Param("user_cluster")
	if userCluster != "" {
		if id, err := utils.UrlSafeBase64Decode(userCluster); err == nil {
			if string(id) != "" {
				// 路径中传了集群名称
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
			klog.V(6).Infof("ClusterRunAnalysis result: %v", result)
			service.ClusterService().SetClusterScanStatus(cfg.ClusterID, result)
		} else {
			klog.V(6).Infof("ClusterRunAnalysis result error: %v", err)
		}
	}()

	amis.WriteJsonOKMsg(c, "后台执行，请稍后查看")
}

// @Summary 获取集群K8sGPT分析结果
// @Security BearerAuth
// @Param cluster query string true "集群名称"
// @Param user_cluster path string true "用户集群标识"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/k8s_gpt/cluster/{user_cluster}/result [get]
func (cc *Controller) GetClusterRunAnalysisResult(c *gin.Context) {
	userCluster := c.Param("user_cluster")
	if userCluster != "" {
		if id, err := utils.UrlSafeBase64Decode(userCluster); err == nil {
			if string(id) != "" {
				// 路径中传了集群名称
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
