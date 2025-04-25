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

func BatchStop(c *gin.Context) {
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
	if err := c.ShouldBindJSON(&req); err != nil {
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
func BatchRestore(c *gin.Context) {
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
	if err := c.ShouldBindJSON(&req); err != nil {
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
func Restart(c *gin.Context) {
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
func BatchRestart(c *gin.Context) {
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
	if err := c.ShouldBindJSON(&req); err != nil {
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
func History(c *gin.Context) {
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
func HistoryRevisionDiff(c *gin.Context) {
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
func Pause(c *gin.Context) {
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
func Resume(c *gin.Context) {
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
func Scale(c *gin.Context) {
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
func Undo(c *gin.Context) {
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

// Event 显示deploy下所有的事件列表，包括deploy、rs、pod
func Event(c *gin.Context) {
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

	var eventList []unstructured.Unstructured

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

func HPA(c *gin.Context) {
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

// 创建deployment
func Create(c *gin.Context) {
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

	if err := c.ShouldBindJSON(&req); err != nil {
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
	deployment := &v1.Deployment{
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
		deployment.Spec.Template.Spec.Containers = append(deployment.Spec.Template.Spec.Containers, corev1.Container{
			Name:            container.Name,
			Image:           container.Image,
			ImagePullPolicy: container.ImagePullPolicy,
		})
	}

	// 创建Deployment
	err = kom.Cluster(selectedCluster).
		WithContext(ctx).
		Resource(deployment).
		Namespace(req.Metadata.Namespace).
		Create(deployment).Error
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOK(c)
}
