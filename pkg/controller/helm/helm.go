package helm

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/weibaohui/k8m/pkg/comm/utils/amis"
	"github.com/weibaohui/k8m/pkg/constants"
	"github.com/weibaohui/k8m/pkg/helm"
	"github.com/weibaohui/k8m/pkg/models"
	"github.com/weibaohui/k8m/pkg/service"
)

func getHelm(c *gin.Context, namespace string) (helm.Helm, error) {
	// if namespace == "" {
	// 	namespace = "default"
	// }
	selectedCluster, err := amis.GetSelectedCluster(c)
	if err != nil {
		amis.WriteJsonError(c, err)
		return nil, err
	}
	restConfig := service.ClusterService().GetClusterByID(selectedCluster).GetRestConfig()
	h, err := helm.New(restConfig, namespace)
	return h, err
}

func handleCommonLogic(c *gin.Context, action string, releaseName, namespace, repoName string) (string, string, error) {
	cluster, _ := amis.GetSelectedCluster(c)
	ctx := amis.GetContextWithUser(c)
	username := fmt.Sprintf("%s", ctx.Value(constants.JwtUserName))
	role := fmt.Sprintf("%s", ctx.Value(constants.JwtUserRole))

	log := models.OperationLog{
		Action:       action,
		Cluster:      cluster,
		Kind:         "Helm",
		Name:         releaseName,
		Namespace:    namespace,
		UserName:     username,
		Group:        repoName,
		Role:         role,
		ActionResult: "success",
	}

	var err error
	if role == constants.RoleClusterReadonly {
		err = fmt.Errorf("非管理员不能%s资源", action)
	}
	if err != nil {
		log.ActionResult = err.Error()
	}
	go func() {
		time.Sleep(1 * time.Second)
		service.OperationLogService().Add(&log)
	}()
	return username, role, err
}
