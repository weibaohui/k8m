package deploy

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/klog/v2"
	"k8s.io/kubectl/pkg/util/deployment"
	"sigs.k8s.io/yaml"
)

type ActionController struct{}

func RegisterActionRoutes(api *gin.RouterGroup) {
	ctrl := &ActionController{}
	api.POST("/deploy/ns/:ns/name/:name/restart", ctrl.Restart)
	api.POST("/deploy/batch/restart", ctrl.BatchRestart)
	api.POST("/deploy/batch/stop", ctrl.BatchStop)
	api.POST("/deploy/batch/restore", ctrl.BatchRestore)
	api.POST("/deploy/ns/:ns/name/:name/revision/:revision/rollout/undo", ctrl.Undo)
	api.GET("/deploy/ns/:ns/name/:name/rollout/history", ctrl.History)
	api.GET("/deploy/ns/:ns/name/:name/revision/:revision/rollout/history", ctrl.HistoryRevisionDiff)
	api.POST("/deploy/ns/:ns/name/:name/rollout/pause", ctrl.Pause)
	api.POST("/deploy/ns/:ns/name/:name/rollout/resume", ctrl.Resume)
	api.POST("/deploy/ns/:ns/name/:name/scale/replica/:replica", ctrl.Scale)
	api.GET("/deploy/ns/:ns/name/:name/events/all", ctrl.Event)
	api.GET("/deploy/ns/:ns/name/:name/hpa", ctrl.HPA)
	api.POST("/deploy/create", ctrl.Create)
	api.POST("/deployment/batch_update_images", ctrl.BatchUpdateImages)

}

// @Summary 批量停止Deployment
// @Security BearerAuth
// @Param cluster path string true "集群名称"
// @Param name_list body []string true "Deployment名称列表"
// @Param ns_list body []string true "命名空间列表"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/deploy/batch/stop [post]
func (nc *ActionController) BatchStop(c *gin.Context) {
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var req struct {
		Names      []string `json:"name_list"`
		Namespaces []string `json:"ns_list"`
	}
	if err = c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	for i := 0; i < len(req.Names); i++ {
		name := req.Names[i]
		ns := req.Namespaces[i]
		x := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Deployment{}).Namespace(ns).Name(name).
			Ctl().Scaler().Stop()
		if x != nil {
			klog.V(6).Infof("批量停止 deploy 错误 %s/%s %v", ns, name, x)

			err = x
		}
	}

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

// @Summary 批量恢复Deployment
// @Security BearerAuth
// @Param cluster path string true "集群名称"
// @Param name_list body []string true "Deployment名称列表"
// @Param ns_list body []string true "命名空间列表"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/deploy/batch/restore [post]
func (nc *ActionController) BatchRestore(c *gin.Context) {
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var req struct {
		Names      []string `json:"name_list"`
		Namespaces []string `json:"ns_list"`
	}
	if err = c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	for i := 0; i < len(req.Names); i++ {
		name := req.Names[i]
		ns := req.Namespaces[i]
		x := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Deployment{}).Namespace(ns).Name(name).
			Ctl().Scaler().Restore()
		if x != nil {
			klog.V(6).Infof("批量恢复 deploy 错误 %s/%s %v", ns, name, x)

			err = x
		}
	}

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

// @Summary 重启单个Deployment
// @Security BearerAuth
// @Param cluster path string true "集群名称"
// @Param ns path string true "命名空间"
// @Param name path string true "Deployment名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/deploy/ns/{ns}/name/{name}/restart [post]
func (nc *ActionController) Restart(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	err = kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Deployment{}).Namespace(ns).Name(name).
		Ctl().Rollout().Restart()
	amis.WriteJsonErrorOrOK(c, err)
}

// @Summary 批量重启Deployment
// @Security BearerAuth
// @Param cluster path string true "集群名称"
// @Param name_list body []string true "Deployment名称列表"
// @Param ns_list body []string true "命名空间列表"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/deploy/batch/restart [post]
func (nc *ActionController) BatchRestart(c *gin.Context) {
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var req struct {
		Names      []string `json:"name_list"`
		Namespaces []string `json:"ns_list"`
	}
	if err = c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	for i := 0; i < len(req.Names); i++ {
		name := req.Names[i]
		ns := req.Namespaces[i]
		x := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Deployment{}).Namespace(ns).Name(name).
			Ctl().Rollout().Restart()
		if x != nil {
			klog.V(6).Infof("批量重启 deploy 错误 %s/%s %v", ns, name, x)

			err = x
		}
	}

	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

// @Summary 获取Deployment历史版本
// @Security BearerAuth
// @Param cluster path string true "集群名称"
// @Param ns path string true "命名空间"
// @Param name path string true "Deployment名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/deploy/ns/{ns}/name/{name}/rollout/history [get]
func (nc *ActionController) History(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	list, err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Deployment{}).Namespace(ns).Name(name).
		Ctl().Rollout().History()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, list)
}

// @Summary 获取Deployment版本差异
// @Security BearerAuth
// @Param cluster path string true "集群名称"
// @Param ns path string true "命名空间"
// @Param name path string true "Deployment名称"
// @Param revision path string true "版本号"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/deploy/ns/{ns}/name/{name}/revision/{revision}/rollout/history [get]
func (nc *ActionController) HistoryRevisionDiff(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	revision := c.Param("revision")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 找到最新的rs
	rsLatest, err := kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.Deployment{}).Namespace(ns).Name(name).
		Ctl().Deployment().ManagedLatestReplicaSet()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	// 找到指定版本的rs
	var rsList []*v1.ReplicaSet
	err = kom.Cluster(selectedCluster).WithContext(ctx).
		Resource(&v1.ReplicaSet{}).Namespace(ns).
		Where(fmt.Sprintf("'metadata.ownerReferences.name'='%s' and 'metadata.ownerReferences.kind'='Deployment'", name)).List(&rsList).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	var rsVersion *v1.ReplicaSet
	for _, r := range rsList {
		if r.ObjectMeta.Annotations != nil && r.ObjectMeta.Annotations[deployment.RevisionAnnotation] == revision {
			rsVersion = r
			break
		}
	}

	current, _ := yaml.JSONToYAML([]byte(utils.ToJSON(rsVersion)))
	latest, _ := yaml.JSONToYAML([]byte(utils.ToJSON(rsLatest)))
	amis.WriteJsonData(c, gin.H{
		"current": string(current),
		"latest":  string(latest),
	})
}

// @Summary 暂停Deployment滚动更新
// @Security BearerAuth
// @Param cluster path string true "集群名称"
// @Param ns path string true "命名空间"
// @Param name path string true "Deployment名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/deploy/ns/{ns}/name/{name}/rollout/pause [post]
func (nc *ActionController) Pause(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	err = kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Deployment{}).Namespace(ns).Name(name).
		Ctl().Rollout().Pause()
	amis.WriteJsonErrorOrOK(c, err)
}

// @Summary 恢复Deployment滚动更新
// @Security BearerAuth
// @Param cluster path string true "集群名称"
// @Param ns path string true "命名空间"
// @Param name path string true "Deployment名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/deploy/ns/{ns}/name/{name}/rollout/resume [post]
func (nc *ActionController) Resume(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	err = kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Deployment{}).Namespace(ns).Name(name).
		Ctl().Rollout().Resume()
	amis.WriteJsonErrorOrOK(c, err)
}

// @Summary 扩缩容Deployment
// @Security BearerAuth
// @Param cluster path string true "集群名称"
// @Param ns path string true "命名空间"
// @Param name path string true "Deployment名称"
// @Param replica path int true "副本数"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/deploy/ns/{ns}/name/{name}/scale/replica/{replica} [post]
func (nc *ActionController) Scale(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	replica := c.Param("replica")
	r := utils.ToInt32(replica)

	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	err = kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Deployment{}).Namespace(ns).Name(name).
		Ctl().Scaler().Scale(r)
	amis.WriteJsonErrorOrOK(c, err)
}

// @Summary 回滚Deployment到指定版本
// @Security BearerAuth
// @Param cluster path string true "集群名称"
// @Param ns path string true "命名空间"
// @Param name path string true "Deployment名称"
// @Param revision path string true "版本号"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/deploy/ns/{ns}/name/{name}/revision/{revision}/rollout/undo [post]
func (nc *ActionController) Undo(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	revision := c.Param("revision")
	ctx := amis.GetContextWithUser(c)
	r := utils.ToInt(revision)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	result, err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Deployment{}).Namespace(ns).Name(name).
		Ctl().Rollout().Undo(r)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOKMsg(c, result)
}

// @Summary 获取Deployment相关事件
// @Security BearerAuth
// @Param cluster path string true "集群名称"
// @Param ns path string true "命名空间"
// @Param name path string true "Deployment名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/deploy/ns/{ns}/name/{name}/events/all [get]
// Event 显示deploy下所有的事件列表，包括deploy、rs、pod
func (nc *ActionController) Event(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var metas []string

	metas = append(metas, name)
	// 先取rs
	rs, err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Deployment{}).
		Namespace(ns).Name(name).
		Ctl().Deployment().ManagedLatestReplicaSet()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	metas = append(metas, rs.ObjectMeta.Name)
	// 再取Pod
	pods, err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Deployment{}).
		Namespace(ns).Name(name).
		Ctl().Deployment().ManagedPods()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	for _, pod := range pods {
		metas = append(metas, pod.ObjectMeta.Name)
	}

	klog.V(6).Infof("meta names = %s", metas)

	var eventList []*unstructured.Unstructured

	sql := kom.Cluster(selectedCluster).
		WithContext(ctx).
		RemoveManagedFields().
		Namespace(ns).
		GVK("events.k8s.io", "v1", "Event")
	// 拼接sql 条件

	// regarding.name = 'x' or regarding.name = 'y'
	var conditions []string
	for _, meta := range metas {
		conditions = append(conditions, fmt.Sprintf("regarding.name = '%s'", meta))
	}
	cc := strings.Join(conditions, " or ")
	if len(metas) > 0 {
		sql = sql.Where(cc)
	}

	err = sql.List(&eventList).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, eventList)
}

// @Summary 获取Deployment的HPA信息
// @Security BearerAuth
// @Param cluster path string true "集群名称"
// @Param ns path string true "命名空间"
// @Param name path string true "Deployment名称"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/deploy/ns/{ns}/name/{name}/hpa [get]
func (nc *ActionController) HPA(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	hpa, err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Deployment{}).Namespace(ns).Name(name).
		Ctl().Deployment().HPAList()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, hpa)
}

// @Summary 创建Deployment
// @Security BearerAuth
// @Param cluster path string true "集群名称"
// @Param body body object true "Deployment配置"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/deploy/create [post]
// 创建deployment
func (nc *ActionController) Create(c *gin.Context) {
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var req struct {
		Metadata struct {
			Namespace   string            `json:"namespace"`
			Name        string            `json:"name"`
			Labels      map[string]string `json:"labels,omitempty"`
			Annotations map[string]string `json:"annotations,omitempty"`
		}
		Spec struct {
			Replicas int32 `json:"replicas"`
			Template struct {
				Spec struct {
					Containers []struct {
						Name            string            `json:"name"`
						Image           string            `json:"image"`
						ImagePullPolicy corev1.PullPolicy `json:"imagePullPolicy"`
					} `json:"containers"`
				} `json:"spec"`
			} `json:"template"`
		}
	}

	if err = c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	// 判断是否存在同名Deployment
	var existingDeployment v1.Deployment
	err = kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Deployment{}).Name(req.Metadata.Name).Namespace(req.Metadata.Namespace).Get(&existingDeployment).Error
	if err == nil {
		amis.WriteJsonError(c, fmt.Errorf("Deployment %s 已存在", req.Metadata.Name))
		return
	}
	// 构建Deployment对象
	item := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:   req.Metadata.Namespace,
			Name:        req.Metadata.Name,
			Labels:      req.Metadata.Labels,
			Annotations: req.Metadata.Annotations,
		},
		Spec: v1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				// 设置标签spec.template.metadata.labels里面的app和version
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":     req.Metadata.Name,
						"version": "v1",
					},
				},
			},
			// 设置标签spec.template.metadata.labels里面的app和version
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app":     req.Metadata.Name,
					"version": "v1",
				},
			},
			// 设置副本数
			Replicas: &req.Spec.Replicas,
		},
	}

	// 设置容器信息
	for _, container := range req.Spec.Template.Spec.Containers {
		item.Spec.Template.Spec.Containers = append(item.Spec.Template.Spec.Containers, corev1.Container{
			Name:            container.Name,
			Image:           container.Image,
			ImagePullPolicy: container.ImagePullPolicy,
		})
	}

	// 创建Deployment
	err = kom.Cluster(selectedCluster).
		WithContext(ctx).
		Resource(item).
		Namespace(req.Metadata.Namespace).
		Create(item).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}

// @Summary 批量更新Deployment镜像
// @Security BearerAuth
// @Param cluster path string true "集群名称"
// @Param deployments body object true "Deployment镜像更新配置"
// @Success 200 {object} string
// @Router /k8s/cluster/{cluster}/deployment/batch_update_images [post]
func (nc *ActionController) BatchUpdateImages(c *gin.Context) {
	ctx := amis.GetContextWithUser(c)
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var req struct {
		Deployments []struct {
			Name       string `json:"name"`
			Namespace  string `json:"namespace"`
			Containers []struct {
				Name  string `json:"name"`
				Image string `json:"image"`
			} `json:"containers"`
		} `json:"deployments"`
	}

	if err = c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	// 用于收集错误和成功信息
	var errors []string
	var successCount int

	// 批量更新每个Deployment的镜像
	for _, deployment := range req.Deployments {
		deploymentKey := fmt.Sprintf("%s/%s", deployment.Namespace, deployment.Name)
		
		// 获取现有的Deployment
		var existingDeployment v1.Deployment
		err = kom.Cluster(selectedCluster).WithContext(ctx).
			Resource(&v1.Deployment{}).
			Namespace(deployment.Namespace).
			Name(deployment.Name).
			Get(&existingDeployment).Error
		if err != nil {
			errorMsg := fmt.Sprintf("Deployment %s 不存在: %v", deploymentKey, err)
			errors = append(errors, errorMsg)
			klog.V(6).Infof("获取Deployment失败 %s: %v", deploymentKey, err)
			continue
		}

		// 标记当前Deployment是否有错误
		hasError := false

		// 更新容器镜像
		for _, containerUpdate := range deployment.Containers {
			found := false
			for i := range existingDeployment.Spec.Template.Spec.Containers {
				if existingDeployment.Spec.Template.Spec.Containers[i].Name == containerUpdate.Name {
					existingDeployment.Spec.Template.Spec.Containers[i].Image = containerUpdate.Image
					found = true
					break
				}
			}
			if !found {
				errorMsg := fmt.Sprintf("容器 %s 在Deployment %s 中未找到", containerUpdate.Name, deploymentKey)
				errors = append(errors, errorMsg)
				klog.V(6).Infof("容器 %s 在Deployment %s 中未找到", containerUpdate.Name, deploymentKey)
				hasError = true
			}
		}

		// 如果容器更新有错误，跳过这个Deployment的更新
		if hasError {
			continue
		}

		// 更新Deployment
		err = kom.Cluster(selectedCluster).WithContext(ctx).
			Resource(&existingDeployment).
			Namespace(deployment.Namespace).
			Update(&existingDeployment).Error
		if err != nil {
			errorMsg := fmt.Sprintf("更新Deployment %s 失败: %v", deploymentKey, err)
			errors = append(errors, errorMsg)
			klog.V(6).Infof("更新Deployment失败 %s: %v", deploymentKey, err)
			continue
		}

		successCount++
		klog.V(6).Infof("成功更新Deployment %s 的镜像", deploymentKey)
	}

	// 构建返回结果
	if len(errors) == 0 {
		// 全部成功
		amis.WriteJsonOKMsg(c, fmt.Sprintf("成功更新 %d 个Deployment的镜像", successCount))
	} else if successCount == 0 {
		// 全部失败
		errorMsg := fmt.Sprintf("批量更新失败，共 %d 个错误：\n%s", len(errors), strings.Join(errors, "\n"))
		amis.WriteJsonError(c, fmt.Errorf(errorMsg))
	} else {
		// 部分成功
		resultMsg := fmt.Sprintf("批量更新完成：成功 %d 个，失败 %d 个。\n失败详情：\n%s", 
			successCount, len(errors), strings.Join(errors, "\n"))
		amis.WriteJsonOKMsg(c, resultMsg)
	}
}
