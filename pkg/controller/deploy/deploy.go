package deploy

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/kom/kom"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
)

func UpdateImageTag(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	selectedCluster := amis.GetSelectedCluster(c)
	ctx := c.Request.Context()

	// json
	// {"container_name":"my-container","image":"my-image","name":"my-container","tag":"sss1","image_pull_secrets":"myregistrykey"}
	type req struct {
		ContainerName    string `json:"container_name"`
		Image            string `json:"image"`
		Tag              string `json:"tag"`
		ImagePullSecrets string `json:"image_pull_secrets"`
	}
	var info req

	if err := c.ShouldBindJSON(&info); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	patchData := `{
  "spec": {
    "template": {
      "spec": {
        "imagePullSecrets": %s,
        "containers": [
          {
            "name": "%s",
            "image": "%s"
          }
        ]
      }
    }
  }
}`
	var json string
	if info.ImagePullSecrets == "" {
		json = "null" // 删除
	} else {
		imagePullSecrets, _ := convertToImagePullSecrets(strings.Split(info.ImagePullSecrets, ","))
		json = utils.ToJSON(imagePullSecrets)
	}

	patchData = fmt.Sprintf(patchData, json, info.ContainerName, info.Image+":"+info.Tag)
	fmt.Println(patchData)
	var item interface{}
	err := kom.Cluster(selectedCluster).
		WithContext(ctx).
		Resource(&v1.Deployment{}).
		Namespace(ns).Name(name).
		Patch(&item, types.MergePatchType, patchData).Error
	amis.WriteJsonErrorOrOK(c, err)
}

// convertToImagePullSecrets converts a []string to JSON format for imagePullSecrets
func convertToImagePullSecrets(secretNames []string) ([]map[string]string, error) {
	// Create a slice of maps
	var result []map[string]string
	for _, name := range secretNames {
		result = append(result, map[string]string{"name": name})
	}
	return result, nil
}
func BatchStop(c *gin.Context) {
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	var req struct {
		Names      []string `json:"name_list"`
		Namespaces []string `json:"ns_list"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var err error
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
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	var req struct {
		Names      []string `json:"name_list"`
		Namespaces []string `json:"ns_list"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var err error
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
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Deployment{}).Namespace(ns).Name(name).
		Ctl().Rollout().Restart()
	amis.WriteJsonErrorOrOK(c, err)
}
func BatchRestart(c *gin.Context) {
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	var req struct {
		Names      []string `json:"name_list"`
		Namespaces []string `json:"ns_list"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		amis.WriteJsonError(c, err)
		return
	}

	var err error
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
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	list, err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Deployment{}).Namespace(ns).Name(name).
		Ctl().Rollout().History()
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonData(c, list)
}
func Pause(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Deployment{}).Namespace(ns).Name(name).
		Ctl().Rollout().Pause()
	amis.WriteJsonErrorOrOK(c, err)
}
func Resume(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Deployment{}).Namespace(ns).Name(name).
		Ctl().Rollout().Resume()
	amis.WriteJsonErrorOrOK(c, err)
}
func Scale(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	replica := c.Param("replica")
	r := utils.ToInt32(replica)

	ctx := c.Request.Context()
	selectedCluster := amis.GetSelectedCluster(c)

	err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Deployment{}).Namespace(ns).Name(name).
		Ctl().Scaler().Scale(r)
	amis.WriteJsonErrorOrOK(c, err)
}
func Undo(c *gin.Context) {
	ns := c.Param("ns")
	name := c.Param("name")
	revision := c.Param("revision")
	ctx := c.Request.Context()
	r := utils.ToInt(revision)
	selectedCluster := amis.GetSelectedCluster(c)

	result, err := kom.Cluster(selectedCluster).WithContext(ctx).Resource(&v1.Deployment{}).Namespace(ns).Name(name).
		Ctl().Rollout().Undo(r)
	if err != nil {
		amis.WriteJsonError(c, err)
		return
	}
	amis.WriteJsonOKMsg(c, result)
}
